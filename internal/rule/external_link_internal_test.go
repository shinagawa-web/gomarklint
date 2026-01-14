package rule

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_checkURL(t *testing.T) {
	tests := []struct {
		name           string
		headStatus     int
		headError      error
		getStatus      int
		getError       error
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "HEAD succeeds",
			headStatus:     200,
			expectedStatus: 200,
			expectError:    false,
		},
		{
			name:           "HEAD fails, GET succeeds",
			headError:      errors.New("HEAD error"),
			getStatus:      404,
			expectedStatus: 404,
			expectError:    false,
		},
		{
			name:        "HEAD and GET both fail",
			headError:   errors.New("HEAD error"),
			getError:    errors.New("GET error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				if callCount == 1 && tt.headError == nil {
					w.WriteHeader(tt.headStatus)
				} else if callCount == 2 && tt.getError == nil {
					w.WriteHeader(tt.getStatus)
				} else {
					// Simulate error by closing the connection early
					hj, ok := w.(http.Hijacker)
					if ok {
						conn, _, _ := hj.Hijack()
						_ = conn.Close()
					}
				}
			}))
			defer ts.Close()

			status, err := checkURL(ts.Client(), ts.URL, 10)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("did not expect error but got: %v", err)
			}
			if !tt.expectError && status != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, status)
			}
		})
	}
}

func Test_performCheck(t *testing.T) {
	t.Run("invalid URL returns error", func(t *testing.T) {
		client := &http.Client{}
		// Use a URL that will cause http.NewRequest to fail
		// URLs with control characters or invalid schemes cause errors
		invalidURLs := []string{
			"ht tp://example.com",     // space in scheme
			"http://exa mple.com",     // space in host
			"http://example.com/\x00", // null byte
			string([]byte{0x7f}),      // invalid characters
		}

		for _, invalidURL := range invalidURLs {
			status, err := performCheck(client, invalidURL)
			if err == nil {
				t.Errorf("expected error for invalid URL %q, but got none", invalidURL)
			}
			if status != 0 {
				t.Errorf("expected status 0 for invalid URL %q, got %d", invalidURL, status)
			}
		}
	})

	t.Run("HEAD success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "HEAD" {
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer ts.Close()

		status, err := performCheck(ts.Client(), ts.URL)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if status != http.StatusOK {
			t.Errorf("expected status 200, got %d", status)
		}
	})

	t.Run("HEAD fails, GET succeeds", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "HEAD" {
				// Close connection to simulate HEAD failure
				hj, ok := w.(http.Hijacker)
				if ok {
					conn, _, _ := hj.Hijack()
					_ = conn.Close()
				}
				return
			}
			if r.Method == "GET" {
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer ts.Close()

		status, err := performCheck(ts.Client(), ts.URL)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if status != http.StatusOK {
			t.Errorf("expected status 200, got %d", status)
		}
	})
}

func Test_formatLinkError(t *testing.T) {
	tests := []struct {
		url      string
		status   int
		err      error
		expected string
	}{
		{
			url:      "https://example.com/404",
			expected: "Link unreachable: https://example.com/404",
		},
		{
			url:      "https://example.com/timeout",
			expected: "Link unreachable: https://example.com/timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			msg := formatLinkError(tt.url)
			if !strings.Contains(msg, tt.expected) && msg != tt.expected {
				t.Errorf("unexpected message:\n got: %s\nwant: %s", msg, tt.expected)
			}
		})
	}
}
