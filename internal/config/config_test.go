package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfig_ValidFile(t *testing.T) {
	json := `{
		"minHeadingLevel": 3,
		"enableLinkCheck": true,
		"skipLinkPatterns": ["localhost", "example.com"]
	}`

	tmp := filepath.Join(t.TempDir(), ".gomarklint.json")
	if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.MinHeadingLevel != 3 {
		t.Errorf("expected minHeadingLevel=3, got %d", cfg.MinHeadingLevel)
	}
	if !cfg.EnableLinkCheck {
		t.Errorf("expected EnableLinkCheck=true")
	}
	if len(cfg.SkipLinkPatterns) != 2 {
		t.Errorf("expected 2 skip patterns, got %d", len(cfg.SkipLinkPatterns))
	}
	// LinkCheckTimeoutSeconds should use default if not specified
	if cfg.LinkCheckTimeoutSeconds != 5 {
		t.Errorf("expected LinkCheckTimeoutSeconds=5 (default), got %d", cfg.LinkCheckTimeoutSeconds)
	}
}

func TestLoadConfig_InvalidField(t *testing.T) {
	json := `{
		"unknown": true
	}`

	tmp := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(tmp)
	if err == nil {
		t.Error("expected error for unknown field, got nil")
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("nonexistent.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	json := `{ invalid json }`

	tmp := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(tmp, []byte(json), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(tmp)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestDefaultConfig(t *testing.T) {
	want := Config{
		MinHeadingLevel:                 2,
		EnableLinkCheck:                 false,
		LinkCheckTimeoutSeconds:         5,
		SkipLinkPatterns:                []string{},
		Include:                         []string{"README.md", "testdata"},
		Ignore:                          []string{},
		OutputFormat:                    "text",
		EnableDuplicateHeadingCheck:     true,
		EnableHeadingLevelCheck:         true,
		EnableNoMultipleBlankLinesCheck: true,
		EnableNoSetextHeadingsCheck:     true,
		EnableFinalBlankLineCheck:       true,
	}

	got := Default()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Default() = %+v, want %+v", got, want)
	}
}
