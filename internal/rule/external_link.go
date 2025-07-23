package rule

import (
	"fmt"
	"net/http"
	"time"

	"github.com/shinagawa-web/gomarklint/internal/parser"
)

func CheckExternalLinks(path string, content string) []LintError {
	urls := parser.ExtractExternalLinks(content)

	var errs []LintError
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, url := range urls {
		status, err := checkURL(client, url)
		if err != nil || status >= 400 {
			errs = append(errs, LintError{
				File:    path,
				Line:    1,
				Message: formatLinkError(url, status, err),
			})
		}
	}

	return errs
}

func checkURL(client *http.Client, url string) (int, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		return resp.StatusCode, nil
	}

	// fallback to GET
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err = client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func formatLinkError(url string, status int, err error) string {
	if err != nil {
		return fmt.Sprintf("Link unreachable: %s (%v)", url, err)
	}
	return fmt.Sprintf("Link unreachable: %s (status: %d)", url, status)
}
