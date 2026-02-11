package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestLoadOrDefault(t *testing.T) {
	t.Run("LoadsExistingConfig", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".gomarklint.json")
		content := `{"minHeadingLevel": 3, "enableLinkCheck": true}`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}

		cfg, err := LoadOrDefault(configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.MinHeadingLevel != 3 {
			t.Errorf("expected MinHeadingLevel=3, got %d", cfg.MinHeadingLevel)
		}
		if !cfg.EnableLinkCheck {
			t.Error("expected EnableLinkCheck=true")
		}
	})

	t.Run("ReturnsDefaultWhenFileDoesNotExist", func(t *testing.T) {
		cfg, err := LoadOrDefault("/nonexistent/path/.gomarklint.json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		defaultCfg := Default()
		if cfg.MinHeadingLevel != defaultCfg.MinHeadingLevel {
			t.Errorf("expected default MinHeadingLevel=%d, got %d", defaultCfg.MinHeadingLevel, cfg.MinHeadingLevel)
		}
	})

	t.Run("ReturnsErrorForPermissionDenied", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".gomarklint.json")
		content := `{"minHeadingLevel": 3}`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}

		// Make directory unreadable to simulate permission error
		if err := os.Chmod(tmpDir, 0000); err != nil {
			t.Fatalf("failed to change permissions: %v", err)
		}
		defer os.Chmod(tmpDir, 0755) // Restore permissions for cleanup

		_, err := LoadOrDefault(configPath)
		if err == nil {
			t.Error("expected error for permission denied, got nil")
		}
	})
}

func TestMergeFlags(t *testing.T) {
	t.Run("MergesChangedFlags", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Int("min-heading", 2, "")
		cmd.Flags().Bool("enable-link-check", false, "")
		cmd.Flags().String("output", "text", "")

		// Simulate setting flags
		_ = cmd.Flags().Set("min-heading", "4")
		_ = cmd.Flags().Set("enable-link-check", "true")

		cfg := Default()
		flags := FlagValues{
			MinHeadingLevel: 4,
			EnableLinkCheck: true,
			OutputFormat:    "text",
		}

		merged := MergeFlags(cfg, cmd, flags)

		if merged.MinHeadingLevel != 4 {
			t.Errorf("expected MinHeadingLevel=4, got %d", merged.MinHeadingLevel)
		}
		if !merged.EnableLinkCheck {
			t.Error("expected EnableLinkCheck=true")
		}
	})

	t.Run("DoesNotMergeUnchangedFlags", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Int("min-heading", 2, "")
		cmd.Flags().String("output", "text", "")

		// Don't set any flags

		cfg := Config{MinHeadingLevel: 3, OutputFormat: "json"}
		flags := FlagValues{
			MinHeadingLevel: 2,
			OutputFormat:    "text",
		}

		merged := MergeFlags(cfg, cmd, flags)

		if merged.MinHeadingLevel != 3 {
			t.Errorf("expected MinHeadingLevel unchanged at 3, got %d", merged.MinHeadingLevel)
		}
		if merged.OutputFormat != "json" {
			t.Errorf("expected OutputFormat unchanged at json, got %s", merged.OutputFormat)
		}
	})

	t.Run("MergesAllFlagTypes", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().Int("min-heading", 2, "")
		cmd.Flags().Bool("enable-link-check", false, "")
		cmd.Flags().Bool("enable-heading-level-check", true, "")
		cmd.Flags().Bool("enable-duplicate-heading-check", true, "")
		cmd.Flags().Bool("enable-no-multiple-blank-lines-check", true, "")
		cmd.Flags().Bool("enable-no-setext-headings-check", true, "")
		cmd.Flags().Bool("enable-final-blank-line-check", true, "")
		cmd.Flags().StringArray("skip-link-patterns", nil, "")
		cmd.Flags().String("output", "text", "")

		_ = cmd.Flags().Set("min-heading", "1")
		_ = cmd.Flags().Set("enable-link-check", "true")
		_ = cmd.Flags().Set("enable-heading-level-check", "false")
		_ = cmd.Flags().Set("enable-duplicate-heading-check", "false")
		_ = cmd.Flags().Set("enable-no-multiple-blank-lines-check", "false")
		_ = cmd.Flags().Set("enable-no-setext-headings-check", "false")
		_ = cmd.Flags().Set("enable-final-blank-line-check", "false")
		_ = cmd.Flags().Set("skip-link-patterns", "https://example.com")
		_ = cmd.Flags().Set("output", "json")

		cfg := Default()
		flags := FlagValues{
			MinHeadingLevel:                 1,
			EnableLinkCheck:                 true,
			EnableHeadingLevelCheck:         false,
			EnableDuplicateHeadingCheck:     false,
			EnableNoMultipleBlankLinesCheck: false,
			EnableNoSetextHeadingsCheck:     false,
			EnableFinalBlankLineCheck:       false,
			SkipLinkPatterns:                []string{"https://example.com"},
			OutputFormat:                    "json",
		}

		merged := MergeFlags(cfg, cmd, flags)

		if merged.MinHeadingLevel != 1 {
			t.Errorf("expected MinHeadingLevel=1, got %d", merged.MinHeadingLevel)
		}
		if !merged.EnableLinkCheck {
			t.Error("expected EnableLinkCheck=true")
		}
		if merged.EnableHeadingLevelCheck {
			t.Error("expected EnableHeadingLevelCheck=false")
		}
		if merged.EnableDuplicateHeadingCheck {
			t.Error("expected EnableDuplicateHeadingCheck=false")
		}
		if merged.EnableNoMultipleBlankLinesCheck {
			t.Error("expected EnableNoMultipleBlankLinesCheck=false")
		}
		if merged.EnableNoSetextHeadingsCheck {
			t.Error("expected EnableNoSetextHeadingsCheck=false")
		}
		if merged.EnableFinalBlankLineCheck {
			t.Error("expected EnableFinalBlankLineCheck=false")
		}
		if len(merged.SkipLinkPatterns) != 1 || merged.SkipLinkPatterns[0] != "https://example.com" {
			t.Errorf("expected SkipLinkPatterns=[https://example.com], got %v", merged.SkipLinkPatterns)
		}
		if merged.OutputFormat != "json" {
			t.Errorf("expected OutputFormat=json, got %s", merged.OutputFormat)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("ValidTextFormat", func(t *testing.T) {
		cfg := Config{OutputFormat: "text"}
		if err := Validate(cfg); err != nil {
			t.Errorf("unexpected error for valid text format: %v", err)
		}
	})

	t.Run("ValidJSONFormat", func(t *testing.T) {
		cfg := Config{OutputFormat: "json"}
		if err := Validate(cfg); err != nil {
			t.Errorf("unexpected error for valid json format: %v", err)
		}
	})

	t.Run("InvalidFormat", func(t *testing.T) {
		cfg := Config{OutputFormat: "xml"}
		if err := Validate(cfg); err == nil {
			t.Error("expected error for invalid format")
		}
	})

	t.Run("EmptyFormat", func(t *testing.T) {
		cfg := Config{OutputFormat: ""}
		if err := Validate(cfg); err == nil {
			t.Error("expected error for empty format")
		}
	})
}
