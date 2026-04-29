package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

const (
	binaryName = "gomarklint-e2e-test"
)

// runTest is a helper function to run the gomarklint binary and return only the output.
// It ignores the exit code since E2E tests check output content, not exit status.
func runTest(t *testing.T, args ...string) []byte {
	output, _ := runTestWithCmd(t, args...)
	return output
}

// assertOutputContains checks if output contains the expected string
func assertOutputContains(t *testing.T, output []byte, expected string) {
	if !bytes.Contains(output, []byte(expected)) {
		t.Errorf("expected output to contain %q, got: %s", expected, output)
	}
}

// assertOutputNotContains checks if output does not contain the unexpected string
func assertOutputNotContains(t *testing.T, output []byte, unexpected string) {
	if bytes.Contains(output, []byte(unexpected)) {
		t.Errorf("expected output to NOT contain %q, got: %s", unexpected, output)
	}
}

// runTestWithCmd runs the gomarklint binary and returns output and exit code.
// This is used when tests need to verify error conditions.
func runTestWithCmd(t *testing.T, args ...string) ([]byte, error) {
	binaryPath := "./" + binaryName

	// Check if binary exists
	if _, err := os.Stat(binaryPath); err != nil {
		t.Fatalf("e2e test binary not found at %s: %v. Did build-e2e run successfully?", binaryName, err)
	}

	cmd := exec.Command(binaryPath)
	cmd.Args = append(cmd.Args, args...)
	output, err := cmd.CombinedOutput()
	return output, err
}

func TestE2E_BasicFunctionality(t *testing.T) {
	t.Run("ValidMarkdown", func(t *testing.T) {
		output := runTest(t, "fixtures/valid.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "No issues found")
	})

	t.Run("InvalidHeadingLevel", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json")

		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}

		assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
		assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
		assertOutputContains(t, output, "First heading should be level 2")
		assertOutputContains(t, output, "found level 1")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "1 issues found")
		assertOutputNotContains(t, output, "[gomarklint error]:")
	})

	t.Run("DuplicateHeadings", func(t *testing.T) {
		output := runTest(t, "fixtures/duplicate_headings.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Errors in fixtures/duplicate_headings.md:")
		assertOutputContains(t, output, "fixtures/duplicate_headings.md:14:")
		assertOutputContains(t, output, "duplicate heading")
		assertOutputContains(t, output, "\"section one\"")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("MultipleBlankLines", func(t *testing.T) {
		output := runTest(t, "fixtures/multiple_blank_lines.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Errors in fixtures/multiple_blank_lines.md:")
		assertOutputContains(t, output, "fixtures/multiple_blank_lines.md:5:")
		assertOutputContains(t, output, "Multiple consecutive blank lines")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("UnclosedCodeBlock", func(t *testing.T) {
		output := runTest(t, "fixtures/unclosed_code_block.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Errors in fixtures/unclosed_code_block.md:")
		assertOutputContains(t, output, "fixtures/unclosed_code_block.md:5:")
		assertOutputContains(t, output, "Unclosed code block")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("LongerClosingFenceIsValid", func(t *testing.T) {
		output := runTest(t, "fixtures/longer_closing_fence.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "Unclosed code block")
	})

	t.Run("ShorterClosingFenceIsUnclosed", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/shorter_closing_fence.md", "--config", ".gomarklint.json")
		if err == nil {
			t.Error("expected non-zero exit code for unclosed code block")
		}
		assertOutputContains(t, output, "Unclosed code block")
	})

	t.Run("EmptyAltText", func(t *testing.T) {
		output := runTest(t, "fixtures/empty_alt_text.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Errors in fixtures/empty_alt_text.md:")
		assertOutputContains(t, output, "fixtures/empty_alt_text.md:5:")
		assertOutputContains(t, output, "image with empty alt text")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("InvalidExternalLink", func(t *testing.T) {
		output := runTest(t, "fixtures/invalid_external_link.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Errors in fixtures/invalid_external_link.md:")
		assertOutputContains(t, output, "fixtures/invalid_external_link.md:14:")
		assertOutputContains(t, output, "Link unreachable")
		assertOutputContains(t, output, "https://this-domain-definitely-does-not-exist-12345.com")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("NoFinalBlankLine", func(t *testing.T) {
		output := runTest(t, "fixtures/no_final_blank_line.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Errors in fixtures/no_final_blank_line.md:")
		assertOutputContains(t, output, "fixtures/no_final_blank_line.md:3:")
		assertOutputContains(t, output, "Missing final blank line")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("FencedCodeWithLanguage", func(t *testing.T) {
		output := runTest(t, "fixtures/fenced_code_with_lang.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "Fenced code block must have a language identifier")
	})

	t.Run("FencedCodeNoLanguage", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/fenced_code_no_lang.md", "--config", ".gomarklint.json")
		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}
		assertOutputContains(t, output, "Errors in fixtures/fenced_code_no_lang.md:")
		assertOutputContains(t, output, "fixtures/fenced_code_no_lang.md:3:")
		assertOutputContains(t, output, "Fenced code block must have a language identifier")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("SingleH1Valid", func(t *testing.T) {
		output := runTest(t, "fixtures/single_h1_valid.md", "--config", "config-single-h1.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "Multiple H1 headings found")
	})

	t.Run("SingleH1Violation", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/single_h1_violation.md", "--config", "config-single-h1.json")
		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}
		assertOutputContains(t, output, "Errors in fixtures/single_h1_violation.md:")
		assertOutputContains(t, output, "fixtures/single_h1_violation.md:5:")
		assertOutputContains(t, output, "Multiple H1 headings found; only one H1 is allowed per file")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("BlanksAroundHeadingsValid", func(t *testing.T) {
		output := runTest(t, "fixtures/blanks_around_headings_valid.md", "--config", "config-blanks-around-headings.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "heading must be preceded by a blank line")
		assertOutputNotContains(t, output, "heading must be followed by a blank line")
	})

	t.Run("BlanksAroundHeadingsViolation", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/blanks_around_headings_violation.md", "--config", "config-blanks-around-headings.json")
		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}
		assertOutputContains(t, output, "Errors in fixtures/blanks_around_headings_violation.md:")
		assertOutputContains(t, output, "fixtures/blanks_around_headings_violation.md:4:")
		assertOutputContains(t, output, "heading must be preceded by a blank line")
		assertOutputContains(t, output, "fixtures/blanks_around_headings_violation.md:7:")
		assertOutputContains(t, output, "heading must be followed by a blank line")
		assertOutputContains(t, output, "3 issues found")
	})

	t.Run("NoBareURLsValid", func(t *testing.T) {
		output := runTest(t, "fixtures/no_bare_urls_valid.md", "--config", "config-no-bare-urls.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "bare URL found")
	})

	t.Run("NoBareURLsHTMLAttributeValid", func(t *testing.T) {
		// URLs inside href/src attributes must not be flagged.
		output := runTest(t, "fixtures/no_bare_urls_valid.md", "--config", "config-no-bare-urls.json")
		assertOutputNotContains(t, output, "fixtures/no_bare_urls_valid.md:9:")
	})

	t.Run("NoBareURLsHTMLCommentValid", func(t *testing.T) {
		// URLs inside single-line HTML comments must not be flagged.
		output := runTest(t, "fixtures/no_bare_urls_valid.md", "--config", "config-no-bare-urls.json")
		assertOutputNotContains(t, output, "fixtures/no_bare_urls_valid.md:7:")
	})

	t.Run("NoBareURLsMultipleCommentsOnLineValid", func(t *testing.T) {
		// URL inside second unclosed comment on same line must not be flagged.
		output := runTest(t, "fixtures/no_bare_urls_valid.md", "--config", "config-no-bare-urls.json")
		assertOutputNotContains(t, output, "fixtures/no_bare_urls_valid.md:11:")
	})

	t.Run("NoBareURLsFenceInsideCommentValid", func(t *testing.T) {
		// Fence opener inside multi-line HTML comment must not be treated as a
		// code block by either no-bare-urls or fenced-code-language.
		output := runTest(t, "fixtures/no_bare_urls_valid.md", "--config", "config-no-bare-urls.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "Fenced code block must have a language identifier")
	})

	t.Run("NoBareURLsLinkCardValid", func(t *testing.T) {
		// A standalone URL on its own line surrounded by blank lines is a link
		// card and must not be flagged.
		output := runTest(t, "fixtures/no_bare_urls_valid.md", "--config", "config-no-bare-urls.json")
		assertOutputNotContains(t, output, "fixtures/no_bare_urls_valid.md:21:")
	})

	t.Run("NoBareURLsViolation", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/no_bare_urls_violation.md", "--config", "config-no-bare-urls.json")
		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}
		assertOutputContains(t, output, "Errors in fixtures/no_bare_urls_violation.md:")
		assertOutputContains(t, output, "fixtures/no_bare_urls_violation.md:3:")
		assertOutputContains(t, output, "bare URL found, use angle brackets or a Markdown link: https://example.com")
		assertOutputContains(t, output, "fixtures/no_bare_urls_violation.md:5:")
		assertOutputContains(t, output, "bare URL found, use angle brackets or a Markdown link: http://other.example.com")
		assertOutputContains(t, output, "2 issues found")
	})

	t.Run("NoEmptyLinksValid", func(t *testing.T) {
		output := runTest(t, "fixtures/no_empty_links_valid.md", "--config", "config-no-empty-links.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "link has empty destination")
	})

	t.Run("NoEmptyLinksViolation", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/no_empty_links_violation.md", "--config", "config-no-empty-links.json")
		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}
		assertOutputContains(t, output, "Errors in fixtures/no_empty_links_violation.md:")
		assertOutputContains(t, output, "fixtures/no_empty_links_violation.md:3:")
		assertOutputContains(t, output, "link has empty destination: [click here]()")
		assertOutputContains(t, output, "fixtures/no_empty_links_violation.md:5:")
		assertOutputContains(t, output, "link has empty destination: [another](#)")
		assertOutputContains(t, output, "fixtures/no_empty_links_violation.md:7:")
		assertOutputContains(t, output, "link has empty destination: ![broken image](<>)")
		assertOutputContains(t, output, "3 issues found")
	})

	t.Run("NoEmphasisAsHeadingValid", func(t *testing.T) {
		output := runTest(t, "fixtures/no_emphasis_as_heading_valid.md", "--config", "config-no-emphasis-as-heading.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "emphasis used as heading")
	})

	t.Run("NoEmphasisAsHeadingViolation", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/no_emphasis_as_heading_violation.md", "--config", "config-no-emphasis-as-heading.json")
		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}
		assertOutputContains(t, output, "Errors in fixtures/no_emphasis_as_heading_violation.md:")
		assertOutputContains(t, output, "fixtures/no_emphasis_as_heading_violation.md:3:")
		assertOutputContains(t, output, "emphasis used as heading, use ATX heading instead: **Section Title**")
		assertOutputContains(t, output, "fixtures/no_emphasis_as_heading_violation.md:5:")
		assertOutputContains(t, output, "emphasis used as heading, use ATX heading instead: _Another Heading_")
		assertOutputContains(t, output, "2 issues found")
	})

	t.Run("BlanksAroundListsValid", func(t *testing.T) {
		output := runTest(t, "fixtures/blanks_around_lists_valid.md", "--config", "config-blanks-around-lists.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "list must be preceded by a blank line")
		assertOutputNotContains(t, output, "list must be followed by a blank line")
	})

	t.Run("BlanksAroundListsNestedValid", func(t *testing.T) {
		// Nested list items must not trigger a false positive between parent and child.
		output := runTest(t, "fixtures/blanks_around_lists_valid.md", "--config", "config-blanks-around-lists.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "fixtures/blanks_around_lists_valid.md:6:")
		assertOutputNotContains(t, output, "fixtures/blanks_around_lists_valid.md:7:")
	})

	t.Run("BlanksAroundListsViolation", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/blanks_around_lists_violation.md", "--config", "config-blanks-around-lists.json")
		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}
		assertOutputContains(t, output, "Errors in fixtures/blanks_around_lists_violation.md:")
		assertOutputContains(t, output, "fixtures/blanks_around_lists_violation.md:4:")
		assertOutputContains(t, output, "list must be preceded by a blank line")
		assertOutputContains(t, output, "fixtures/blanks_around_lists_violation.md:6:")
		assertOutputContains(t, output, "list must be followed by a blank line")
		assertOutputContains(t, output, "2 issues found")
	})

	t.Run("BlanksAroundFencesValid", func(t *testing.T) {
		output := runTest(t, "fixtures/blanks_around_fences_valid.md", "--config", "config-blanks-around-fences.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "fenced code block must be preceded by a blank line")
		assertOutputNotContains(t, output, "fenced code block must be followed by a blank line")
	})

	t.Run("BlanksAroundFencesViolation", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/blanks_around_fences_violation.md", "--config", "config-blanks-around-fences.json")
		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}
		assertOutputContains(t, output, "Errors in fixtures/blanks_around_fences_violation.md:")
		assertOutputContains(t, output, "fixtures/blanks_around_fences_violation.md:4:")
		assertOutputContains(t, output, "fenced code block must be preceded by a blank line")
		assertOutputContains(t, output, "fixtures/blanks_around_fences_violation.md:8:")
		assertOutputContains(t, output, "fenced code block must be followed by a blank line")
		assertOutputContains(t, output, "2 issues found")
	})

	t.Run("MaxLineLengthValid", func(t *testing.T) {
		output := runTest(t, "fixtures/max_line_length_valid.md", "--config", "config-max-line-length.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "line exceeds")
	})

	t.Run("MaxLineLengthViolation", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/max_line_length_violation.md", "--config", "config-max-line-length.json")
		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}
		assertOutputContains(t, output, "Errors in fixtures/max_line_length_violation.md:")
		assertOutputContains(t, output, "fixtures/max_line_length_violation.md:5:")
		assertOutputContains(t, output, "line exceeds 80 bytes (100)")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("NoHardTabsValid", func(t *testing.T) {
		output := runTest(t, "fixtures/no_hard_tabs_valid.md", "--config", "config-no-hard-tabs.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "hard tab character")
	})

	t.Run("NoHardTabsViolation", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/no_hard_tabs_violation.md", "--config", "config-no-hard-tabs.json")
		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}
		assertOutputContains(t, output, "Errors in fixtures/no_hard_tabs_violation.md:")
		assertOutputContains(t, output, "fixtures/no_hard_tabs_violation.md:3:")
		assertOutputContains(t, output, "hard tab character found at column 1")
		assertOutputContains(t, output, "fixtures/no_hard_tabs_violation.md:5:")
		assertOutputContains(t, output, "hard tab character found at column 4")
		assertOutputContains(t, output, "fixtures/no_hard_tabs_violation.md:7:")
		assertOutputContains(t, output, "3 issues found")
	})

	t.Run("NoTrailingPunctuationValid", func(t *testing.T) {
		output := runTest(t, "fixtures/no_trailing_punctuation_valid.md", "--config", "config-no-trailing-punctuation.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "no-trailing-punctuation")
	})

	t.Run("NoTrailingPunctuationViolation", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/no_trailing_punctuation_violation.md", "--config", "config-no-trailing-punctuation.json")
		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}
		assertOutputContains(t, output, "Errors in fixtures/no_trailing_punctuation_violation.md:")
		assertOutputContains(t, output, "fixtures/no_trailing_punctuation_violation.md:1:")
		assertOutputContains(t, output, `heading ends with "."`)
		assertOutputContains(t, output, "fixtures/no_trailing_punctuation_violation.md:5:")
		assertOutputContains(t, output, `heading ends with ","`)
		assertOutputContains(t, output, "fixtures/no_trailing_punctuation_violation.md:9:")
		assertOutputContains(t, output, `heading ends with "!"`)
		assertOutputContains(t, output, "3 issues found")
	})

	t.Run("LinkFragmentsValid", func(t *testing.T) {
		output := runTest(t, "fixtures/link_fragments_valid.md", "--config", "config-link-fragments.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "link-fragments")
	})

	t.Run("LinkFragmentsViolation", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/link_fragments_violation.md", "--config", "config-link-fragments.json")
		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}
		assertOutputContains(t, output, "Errors in fixtures/link_fragments_violation.md:")
		assertOutputContains(t, output, "fixtures/link_fragments_violation.md:3:")
		assertOutputContains(t, output, "link-fragments")
		assertOutputContains(t, output, "#setup")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("LinkFragmentsCustomValid", func(t *testing.T) {
		output := runTest(t, "fixtures/link_fragments_custom_valid.md", "--config", "config-link-fragments-custom.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "link-fragments")
	})

	t.Run("LinkFragmentsCustomViolation", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/link_fragments_custom_violation.md", "--config", "config-link-fragments-custom.json")
		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}
		assertOutputContains(t, output, "Errors in fixtures/link_fragments_custom_violation.md:")
		assertOutputContains(t, output, "fixtures/link_fragments_custom_violation.md:3:")
		assertOutputContains(t, output, "link-fragments")
		assertOutputContains(t, output, "#setup")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("ConsistentCodeFenceValid", func(t *testing.T) {
		output := runTest(t, "fixtures/consistent_code_fence_valid.md", "--config", "config-consistent-code-fence.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "consistent-code-fence")
	})

	t.Run("ConsistentCodeFenceViolation", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/consistent_code_fence_violation.md", "--config", "config-consistent-code-fence.json")
		if err == nil {
			t.Error("expected non-zero exit code for lint violations")
		}
		assertOutputContains(t, output, "Errors in fixtures/consistent_code_fence_violation.md:")
		assertOutputContains(t, output, "fixtures/consistent_code_fence_violation.md:7:")
		assertOutputContains(t, output, "expected '```' fence, got '~~~' fence")
		assertOutputContains(t, output, "1 issues found")
	})
}

func TestE2E_Configuration(t *testing.T) {
	t.Run("MinHeadingLevel1", func(t *testing.T) {
		output := runTest(t, "fixtures/heading_level_one.md", "--config", "config-min-heading-1.json")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "First heading should be level")
		assertOutputNotContains(t, output, "Errors")
	})

	t.Run("MinHeadingLevel1EnabledOmitted", func(t *testing.T) {
		output := runTest(t, "fixtures/heading_level_one.md", "--config", "config-min-heading-1-no-enabled.json")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "First heading should be level")
		assertOutputNotContains(t, output, "Errors")
	})

	t.Run("DisableDuplicateHeadingRule", func(t *testing.T) {
		output := runTest(t, "fixtures/duplicate_headings.md", "--config", "config-no-duplicate-heading.json")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "duplicate heading")
		assertOutputNotContains(t, output, "Errors")
	})

	t.Run("OffValueDisablesRule", func(t *testing.T) {
		output := runTest(t, "fixtures/duplicate_headings.md", "--config", "config-off-duplicate-heading.json")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "duplicate heading")
		assertOutputNotContains(t, output, "Errors")
	})

	t.Run("DefaultFalseOptInModeWithViolation", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/no_final_blank_line.md", "--config", "config-opt-in.json")
		if err == nil {
			t.Error("expected exit 1: opt-in rule final-blank-line should detect violation")
		}
		assertOutputContains(t, output, "Missing final blank line")
		assertOutputNotContains(t, output, "Setext heading found")
	})

	t.Run("DisableFinalBlankLineRule", func(t *testing.T) {
		output := runTest(t, "fixtures/no_final_blank_line.md", "--config", "config-no-final-blank-line.json")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "Missing final blank line")
		assertOutputNotContains(t, output, "Errors")
	})
}

func TestE2E_OutputFormats(t *testing.T) {
	t.Run("TextFormat", func(t *testing.T) {
		output := runTest(t, "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json", "--output", "text")
		assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
		assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
		assertOutputContains(t, output, "First heading should be level 2")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("JSONFormat", func(t *testing.T) {
		output := runTest(t, "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json", "--output", "json")
		assertOutputContains(t, output, `"files"`)
		assertOutputContains(t, output, `"total"`)
		assertOutputContains(t, output, `"details"`)
		assertOutputContains(t, output, `"elapsed_ms"`)
		assertOutputContains(t, output, `"file": "fixtures/invalid_heading_level.md"`)
		assertOutputContains(t, output, `"line": 1`)
		assertOutputContains(t, output, `"message": "First heading should be level 2`)
		assertOutputContains(t, output, `{`)
		assertOutputContains(t, output, `}`)
	})
}

func TestE2E_MultipleFiles(t *testing.T) {
	t.Run("MultipleFiles", func(t *testing.T) {
		output := runTest(t, "fixtures/valid.md", "fixtures/invalid_heading_level.md", "--config", ".gomarklint.json")
		assertOutputNotContains(t, output, "Errors in fixtures/valid.md")
		assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
		assertOutputContains(t, output, "fixtures/invalid_heading_level.md:1:")
		assertOutputContains(t, output, "First heading should be level 2")
		assertOutputContains(t, output, "Checked 2 file(s)")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("DirectoryRecursion", func(t *testing.T) {
		output := runTest(t, "fixtures", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
		assertOutputContains(t, output, "First heading should be level 2 (found level 1)")
		assertOutputContains(t, output, "Errors in fixtures/duplicate_headings.md:")
		assertOutputContains(t, output, "duplicate heading: \"section one\"")
		assertOutputContains(t, output, "Errors in fixtures/multiple_blank_lines.md:")
		assertOutputContains(t, output, "Multiple consecutive blank lines")
		assertOutputContains(t, output, "Errors in fixtures/unclosed_code_block.md:")
		assertOutputContains(t, output, "Unclosed code block")
		assertOutputContains(t, output, "Errors in fixtures/empty_alt_text.md:")
		assertOutputContains(t, output, "image with empty alt text")
		assertOutputContains(t, output, "Errors in fixtures/invalid_external_link.md:")
		assertOutputContains(t, output, "Link unreachable: https://this-domain-definitely-does-not-exist-12345.com")
		assertOutputContains(t, output, "Errors in fixtures/multiple_external_links.md:")
		assertOutputContains(t, output, "this-is-definitely-an-invalid-domain-12345.xyz")
		assertOutputContains(t, output, "another-invalid-domain-67890.test")
		assertOutputContains(t, output, "Errors in fixtures/heading_level_one.md:")
		assertOutputContains(t, output, "Errors in fixtures/empty.md:")
		assertOutputContains(t, output, "Missing final blank line")
		assertOutputContains(t, output, "Errors in fixtures/multiple_violations.md:")
		assertOutputContains(t, output, "Errors in fixtures/setext_headings.md:")
		assertOutputContains(t, output, "Setext heading found")
		assertOutputContains(t, output, "Errors in fixtures/mixed_severity.md:")
		assertOutputContains(t, output, "Unclosed code block")
		assertOutputContains(t, output, "Errors in fixtures/single_h1_violation.md:")
		assertOutputContains(t, output, "Multiple H1 headings found")
		assertOutputContains(t, output, "Errors in fixtures/single_h1_valid.md:")
		assertOutputContains(t, output, "Errors in fixtures/shorter_closing_fence.md:")
		assertOutputContains(t, output, "Errors in fixtures/blanks_around_headings_violation.md:")
		assertOutputContains(t, output, "heading must be preceded by a blank line")
		assertOutputContains(t, output, "Errors in fixtures/no_bare_urls_violation.md:")
		assertOutputContains(t, output, "bare URL found")
		assertOutputContains(t, output, "Errors in fixtures/no_empty_links_violation.md:")
		assertOutputContains(t, output, "link has empty destination")
		assertOutputContains(t, output, "Errors in fixtures/no_emphasis_as_heading_violation.md:")
		assertOutputContains(t, output, "emphasis used as heading")
		assertOutputContains(t, output, "Errors in fixtures/blanks_around_lists_violation.md:")
		assertOutputContains(t, output, "list must be preceded by a blank line")
		assertOutputContains(t, output, "Errors in fixtures/no_hard_tabs_violation.md:")
		assertOutputContains(t, output, "hard tab character found")
		assertOutputContains(t, output, "Errors in fixtures/blanks_around_fences_violation.md:")
		assertOutputContains(t, output, "fenced code block must be preceded by a blank line")
		assertOutputContains(t, output, "Errors in fixtures/consistent_code_fence_violation.md:")
		assertOutputContains(t, output, "expected '```' fence, got '~~~' fence")
		assertOutputContains(t, output, "Checked 51 file(s)")
		assertOutputNotContains(t, output, "Errors in fixtures/valid.md")
		assertOutputNotContains(t, output, "Errors in fixtures/with_frontmatter.md")
		assertOutputNotContains(t, output, "Errors in fixtures/frontmatter_only.md")
		assertOutputNotContains(t, output, "Errors in fixtures/valid_external_links.md")
		assertOutputNotContains(t, output, "Errors in fixtures/mixed_link_types.md")
		assertOutputNotContains(t, output, "Errors in fixtures/longer_closing_fence.md")
		assertOutputNotContains(t, output, "Errors in fixtures/blanks_around_headings_valid.md:")
		assertOutputNotContains(t, output, "Errors in fixtures/no_bare_urls_valid.md:")
		assertOutputNotContains(t, output, "Errors in fixtures/blanks_around_lists_valid.md:")
		assertOutputNotContains(t, output, "Errors in fixtures/consistent_code_fence_valid.md:")
		assertOutputNotContains(t, output, "Errors in fixtures/no_empty_links_valid.md:")
		assertOutputNotContains(t, output, "Errors in fixtures/no_emphasis_as_heading_valid.md:")
		assertOutputNotContains(t, output, "Errors in fixtures/no_hard_tabs_valid.md:")
		assertOutputNotContains(t, output, "Errors in fixtures/blanks_around_fences_valid.md:")
	})

	t.Run("ErrorsFromAllFiles", func(t *testing.T) {
		output := runTest(t, "fixtures/invalid_heading_level.md", "fixtures/duplicate_headings.md", "fixtures/multiple_blank_lines.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Errors in fixtures/invalid_heading_level.md:")
		assertOutputContains(t, output, "First heading should be level 2")
		assertOutputContains(t, output, "Errors in fixtures/duplicate_headings.md:")
		assertOutputContains(t, output, "duplicate heading")
		assertOutputContains(t, output, "Errors in fixtures/multiple_blank_lines.md:")
		assertOutputContains(t, output, "Multiple consecutive blank lines")
		assertOutputContains(t, output, "Checked 3 file(s)")
		assertOutputContains(t, output, "3 issues found")
	})
}

func TestE2E_EdgeCases(t *testing.T) {
	t.Run("NonExistentFile", func(t *testing.T) {
		output := runTest(t, "fixtures/nonexistent.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Checked 0 file(s)")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "Errors")
	})

	t.Run("InvalidConfigFile", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/valid.md", "--config", "invalid.json")
		if err == nil {
			t.Errorf("expected error for invalid config file, but command succeeded")
		}
		assertOutputContains(t, output, "[gomarklint error]:")
		assertOutputContains(t, output, "failed to parse config file")
	})

	t.Run("EmptyFile", func(t *testing.T) {
		output := runTest(t, "fixtures/empty.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Errors in fixtures/empty.md:")
		assertOutputContains(t, output, "fixtures/empty.md:1:")
		assertOutputContains(t, output, "Missing final blank line")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("FilesWithFrontmatter", func(t *testing.T) {
		output := runTest(t, "fixtures/with_frontmatter.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "Errors")
	})

	t.Run("FrontmatterOnlyNoFalsePositive", func(t *testing.T) {
		output := runTest(t, "fixtures/frontmatter_only.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "Missing final blank line")
	})

	t.Run("MultipleViolationsInSingleFile", func(t *testing.T) {
		output, _ := runTestWithCmd(t, "fixtures/multiple_violations.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Errors in fixtures/multiple_violations.md:")
		assertOutputContains(t, output, "fixtures/multiple_violations.md:6:")
		assertOutputContains(t, output, "First heading should be level 2")
		assertOutputContains(t, output, "fixtures/multiple_violations.md:10:")
		assertOutputContains(t, output, "Multiple consecutive blank lines")
		assertOutputContains(t, output, "fixtures/multiple_violations.md:17:")
		assertOutputContains(t, output, "duplicate heading")
		assertOutputContains(t, output, "section one")
		assertOutputContains(t, output, "fixtures/multiple_violations.md:21:")
		assertOutputContains(t, output, "image with empty alt text")
		assertOutputContains(t, output, "fixtures/multiple_violations.md:25:")
		assertOutputContains(t, output, "Unclosed code block")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "6 issues found")
	})
}

func TestE2E_ExternalLinkChecks(t *testing.T) {
	t.Run("DisableExternalLinkCheck", func(t *testing.T) {
		output := runTest(t, "fixtures/invalid_external_link.md", "--config", "config-no-link-check.json")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "Link unreachable")
	})

	t.Run("EnableExternalLinkCheckWithInvalidLink", func(t *testing.T) {
		output := runTest(t, "fixtures/invalid_external_link.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Errors in fixtures/invalid_external_link.md:")
		assertOutputContains(t, output, "Link unreachable")
		assertOutputContains(t, output, "this-domain-definitely-does-not-exist-12345.com")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("ValidExternalLinksOnly", func(t *testing.T) {
		output := runTest(t, "fixtures/valid_external_links.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "Link unreachable")
	})

	t.Run("MultipleExternalLinks", func(t *testing.T) {
		output := runTest(t, "fixtures/multiple_external_links.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Errors in fixtures/multiple_external_links.md:")
		assertOutputContains(t, output, "Link unreachable")
		assertOutputContains(t, output, "this-is-definitely-an-invalid-domain-12345.xyz")
		assertOutputContains(t, output, "another-invalid-domain-67890.test")
		assertOutputContains(t, output, "2 issues found")
	})

	t.Run("ExternalLinkCheckEnabledByDefault", func(t *testing.T) {
		output := runTest(t, "fixtures/invalid_external_link.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Errors in fixtures/invalid_external_link.md:")
		assertOutputContains(t, output, "Link unreachable")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("HTTPAndHTTPSLinks", func(t *testing.T) {
		output := runTest(t, "fixtures/http_and_https_links.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "link(s)")
	})

	t.Run("MixedLinkTypes", func(t *testing.T) {
		output := runTest(t, "fixtures/mixed_link_types.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "link(s)")
		assertOutputContains(t, output, "No issues found")
	})

	t.Run("SameLineMultipleLinks", func(t *testing.T) {
		output := runTest(t, "fixtures/same_line_multiple_links.md", "--config", ".gomarklint.json")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "No issues found")
		assertOutputContains(t, output, "link(s)")
	})

	t.Run("SkipPatternsExcludeLinks", func(t *testing.T) {
		output := runTest(t, "fixtures/invalid_external_link.md", "--config", "config-skip-patterns.json")
		assertOutputContains(t, output, "Checked 1 file(s)")
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "Link unreachable")
	})
}

func TestE2E_Severity(t *testing.T) {
	t.Run("NoSetextHeadingsBasic", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/setext_headings.md", "--config", ".gomarklint.json")
		if err == nil {
			t.Error("expected non-zero exit code for setext heading violation")
		}
		assertOutputContains(t, output, "Errors in fixtures/setext_headings.md:")
		assertOutputContains(t, output, "fixtures/setext_headings.md:6:")
		assertOutputContains(t, output, "Setext heading found")
		assertOutputContains(t, output, "1 issues found")
	})

	t.Run("WarningSeverityExits0", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/setext_headings.md", "--config", "config-warning-setext.json")
		if err != nil {
			t.Errorf("expected exit 0 for warning-only violations, got error: %v\noutput: %s", err, output)
		}
		assertOutputContains(t, output, "Warnings in fixtures/setext_headings.md:")
		assertOutputContains(t, output, "[warning]")
		assertOutputContains(t, output, "Setext heading found")
		assertOutputContains(t, output, "1 warning found")
	})

	t.Run("ErrorSeverityExits1", func(t *testing.T) {
		_, err := runTestWithCmd(t, "fixtures/setext_headings.md", "--config", ".gomarklint.json")
		if err == nil {
			t.Error("expected exit 1 for error-severity violation")
		}
	})

	t.Run("MixedSeverityExits1", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/mixed_severity.md", "--config", "config-mixed-severity.json")
		if err == nil {
			t.Error("expected exit 1 when at least one error-severity violation exists")
		}
		assertOutputContains(t, output, "[warning]")
		assertOutputContains(t, output, "[error]")
		assertOutputContains(t, output, "Setext heading found")
		assertOutputContains(t, output, "Unclosed code block")
	})

	t.Run("WarningsOnlyExits0", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/setext_headings.md", "--config", "config-warning-setext.json")
		if err != nil {
			t.Errorf("expected exit 0 for warnings-only result, got: %v\noutput: %s", err, output)
		}
		assertOutputContains(t, output, "warning found")
		assertOutputNotContains(t, output, "issues found")
	})

	t.Run("SeverityErrorFlagSuppressesWarnings", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/setext_headings.md", "--config", "config-warning-setext.json", "--severity", "error")
		if err != nil {
			t.Errorf("expected exit 0 when warnings filtered out, got: %v\noutput: %s", err, output)
		}
		assertOutputNotContains(t, output, "[warning]")
		assertOutputNotContains(t, output, "Setext heading found")
		assertOutputContains(t, output, "No issues found")
	})

	t.Run("SeverityWarningFlagShowsAll", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/mixed_severity.md", "--config", "config-mixed-severity.json", "--severity", "warning")
		if err == nil {
			t.Error("expected exit 1 when error violations present")
		}
		assertOutputContains(t, output, "[warning]")
		assertOutputContains(t, output, "[error]")
	})

	t.Run("SeverityErrorFlagMixedFiltersWarnings", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/mixed_severity.md", "--config", "config-mixed-severity.json", "--severity", "error")
		if err == nil {
			t.Error("expected exit 1: error-severity violation remains after filtering warnings")
		}
		assertOutputNotContains(t, output, "[warning]")
		assertOutputNotContains(t, output, "Setext heading found")
		assertOutputContains(t, output, "[error]")
		assertOutputContains(t, output, "Unclosed code block")
	})

	t.Run("DefaultFalseOptInMode", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/setext_headings.md", "--config", "config-opt-in.json")
		if err != nil {
			t.Errorf("expected exit 0 in opt-in mode (only final-blank-line enabled), got: %v\noutput: %s", err, output)
		}
		assertOutputContains(t, output, "No issues found")
		assertOutputNotContains(t, output, "Setext heading found")
	})
}

func TestE2E_DisableComments(t *testing.T) {
	// fixture line map:
	//  4  suppressed: <!-- gomarklint-disable -->
	//  6  reported:   after <!-- gomarklint-enable -->
	//  9  suppressed: <!-- gomarklint-disable no-bare-urls -->
	// 11  reported:   after <!-- gomarklint-enable no-bare-urls -->
	// 13  suppressed: <!-- gomarklint-disable-line -->
	// 15  suppressed: <!-- gomarklint-disable-line no-bare-urls -->
	// 18  suppressed: <!-- gomarklint-disable-next-line -->
	// 21  suppressed: <!-- gomarklint-disable-next-line no-bare-urls -->
	// 23  reported:   typo rule name "no-bare-url" (missing trailing 's') — disable has no effect
	// 26  reported:   nonexistent rule "nonexistent-rule" — disable has no effect

	t.Run("BlockDisableAll", func(t *testing.T) {
		output := runTest(t, "fixtures/disable_comment.md", "--config", "config-no-bare-urls.json")
		assertOutputNotContains(t, output, "fixtures/disable_comment.md:4:")
	})

	t.Run("EnableAll", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/disable_comment.md", "--config", "config-no-bare-urls.json")
		if err == nil {
			t.Error("expected exit 1: violations after enable should be reported")
		}
		assertOutputContains(t, output, "fixtures/disable_comment.md:6:")
	})

	t.Run("BlockDisableNamedRule", func(t *testing.T) {
		output := runTest(t, "fixtures/disable_comment.md", "--config", "config-no-bare-urls.json")
		assertOutputNotContains(t, output, "fixtures/disable_comment.md:9:")
	})

	t.Run("EnableNamedRule", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/disable_comment.md", "--config", "config-no-bare-urls.json")
		if err == nil {
			t.Error("expected exit 1: violations after enable should be reported")
		}
		assertOutputContains(t, output, "fixtures/disable_comment.md:11:")
	})

	t.Run("DisableLineAll", func(t *testing.T) {
		output := runTest(t, "fixtures/disable_comment.md", "--config", "config-no-bare-urls.json")
		assertOutputNotContains(t, output, "fixtures/disable_comment.md:13:")
	})

	t.Run("DisableLineNamedRule", func(t *testing.T) {
		output := runTest(t, "fixtures/disable_comment.md", "--config", "config-no-bare-urls.json")
		assertOutputNotContains(t, output, "fixtures/disable_comment.md:15:")
	})

	t.Run("DisableNextLineAll", func(t *testing.T) {
		output := runTest(t, "fixtures/disable_comment.md", "--config", "config-no-bare-urls.json")
		assertOutputNotContains(t, output, "fixtures/disable_comment.md:18:")
	})

	t.Run("DisableNextLineNamedRule", func(t *testing.T) {
		output := runTest(t, "fixtures/disable_comment.md", "--config", "config-no-bare-urls.json")
		assertOutputNotContains(t, output, "fixtures/disable_comment.md:21:")
	})

	t.Run("WrongRuleName", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/disable_comment.md", "--config", "config-no-bare-urls.json")
		if err == nil {
			t.Error("expected exit 1: typo rule name should not suppress the violation")
		}
		assertOutputContains(t, output, "fixtures/disable_comment.md:23:")
	})

	t.Run("NonexistentRule", func(t *testing.T) {
		output, err := runTestWithCmd(t, "fixtures/disable_comment.md", "--config", "config-no-bare-urls.json")
		if err == nil {
			t.Error("expected exit 1: nonexistent rule name should not suppress the violation")
		}
		assertOutputContains(t, output, "fixtures/disable_comment.md:26:")
	})
}
