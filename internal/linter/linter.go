package linter

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/shinagawa-web/gomarklint/internal/config"
	"github.com/shinagawa-web/gomarklint/internal/file"
	"github.com/shinagawa-web/gomarklint/internal/rule"
)

// Linter performs linting on markdown files.
type Linter struct {
	config           config.Config
	compiledPatterns []*regexp.Regexp
	urlCache         *sync.Map
}

// Result holds the results of a linting run.
type Result struct {
	Errors            map[string][]rule.LintError // Errors per file path
	OrderedPaths      []string                    // Sorted file paths
	TotalErrors       int                         // Total number of errors
	TotalLines        int                         // Total number of lines checked
	TotalLinksChecked int                         // Total number of links checked
	FailedFiles       map[string]error            // Files that failed to read
}

// New creates a new Linter with the given configuration.
func New(cfg config.Config) (*Linter, error) {
	compiledPatterns := []*regexp.Regexp{}
	if cfg.EnableLinkCheck {
		for _, pat := range cfg.SkipLinkPatterns {
			re, err := regexp.Compile(pat)
			if err != nil {
				log.Printf("Invalid skip-link-pattern: %s (error: %v)", pat, err)
				continue
			}
			compiledPatterns = append(compiledPatterns, re)
		}
	}

	return &Linter{
		config:           cfg,
		compiledPatterns: compiledPatterns,
		urlCache:         &sync.Map{},
	}, nil
}

// Run performs linting on the given file paths concurrently.
func (l *Linter) Run(filePaths []string) (*Result, error) {
	// Deduplicate file paths to prevent double-counting
	uniquePaths := make(map[string]struct{})
	for _, p := range filePaths {
		uniquePaths[p] = struct{}{}
	}
	deduped := make([]string, 0, len(uniquePaths))
	for p := range uniquePaths {
		deduped = append(deduped, p)
	}

	results := map[string][]rule.LintError{}
	orderedPaths := make([]string, 0, len(deduped))
	failedFiles := map[string]error{}
	totalErrors := 0
	totalLines := 0
	totalLinksChecked := 0

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, path := range deduped {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			content, err := file.ReadFile(p)
			if err != nil {
				mu.Lock()
				failedFiles[p] = err
				mu.Unlock()
				fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", p, err)
				return
			}
			errors, lineCount, linksChecked := l.collectErrors(p, content)

			mu.Lock()
			results[p] = errors
			orderedPaths = append(orderedPaths, p)
			totalErrors += len(errors)
			totalLines += lineCount
			totalLinksChecked += linksChecked
			mu.Unlock()
		}(path)
	}

	wg.Wait()

	// Sort paths to ensure consistent output order
	sort.Strings(orderedPaths)

	return &Result{
		Errors:            results,
		OrderedPaths:      orderedPaths,
		TotalErrors:       totalErrors,
		TotalLines:        totalLines,
		TotalLinksChecked: totalLinksChecked,
		FailedFiles:       failedFiles,
	}, nil
}

// LintContent performs linting checks on the provided content string.
// This is useful for benchmarking and testing without file I/O overhead.
func (l *Linter) LintContent(path string, content string) ([]rule.LintError, int, int) {
	return l.collectErrors(path, content)
}

// collectErrors performs linting checks on a single file's content.
func (l *Linter) collectErrors(path string, content string) ([]rule.LintError, int, int) {
	body, offset := file.StripFrontmatter(content)
	lines := strings.Split(body, "\n")

	var allErrors []rule.LintError
	if l.config.EnableFinalBlankLineCheck {
		allErrors = append(allErrors, rule.CheckFinalBlankLine(path, lines, offset)...)
	}
	allErrors = append(allErrors, rule.CheckUnclosedCodeBlocks(path, lines, offset)...)
	allErrors = append(allErrors, rule.CheckEmptyAltText(path, lines, offset)...)
	if l.config.EnableHeadingLevelCheck {
		allErrors = append(allErrors, rule.CheckHeadingLevels(path, lines, offset, l.config.MinHeadingLevel)...)
	}
	if l.config.EnableDuplicateHeadingCheck {
		allErrors = append(allErrors, rule.CheckDuplicateHeadings(path, lines, offset)...)
	}
	if l.config.EnableNoMultipleBlankLinesCheck {
		allErrors = append(allErrors, rule.CheckNoMultipleBlankLines(path, lines, offset)...)
	}

	if l.config.EnableNoSetextHeadingsCheck {
		allErrors = append(allErrors, rule.CheckNoSetextHeadings(path, lines, offset)...)
	}

	linksChecked := 0
	if l.config.EnableLinkCheck {
		errors, count := rule.CheckExternalLinks(path, lines, offset, l.compiledPatterns, l.config.LinkCheckTimeoutSeconds, rule.DefaultRetryDelayMs, l.urlCache)
		allErrors = append(allErrors, errors...)
		linksChecked = count
	}

	sort.Slice(allErrors, func(i, j int) bool {
		return allErrors[i].Line < allErrors[j].Line
	})

	lineCount := strings.Count(content, "\n") + 1
	return allErrors, lineCount, linksChecked
}
