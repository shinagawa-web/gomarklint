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
		content := `{
			"default": true,
			"rules": {
				"heading-level": { "enabled": true, "minLevel": 3 },
				"external-link": true
			},
			"output": "json"
		}`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}

		cfg, err := LoadOrDefault(configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.OutputFormat != "json" {
			t.Errorf("expected OutputFormat=json, got %s", cfg.OutputFormat)
		}
		if !cfg.IsEnabled("external-link") {
			t.Error("expected external-link enabled")
		}
		if v, ok := cfg.RuleOptions("heading-level")["minLevel"]; !ok || v.(float64) != 3 {
			t.Errorf("expected heading-level minLevel=3, got %v", v)
		}
	})

	t.Run("ReturnsDefaultWhenFileDoesNotExist", func(t *testing.T) {
		cfg, err := LoadOrDefault("/nonexistent/path/.gomarklint.json")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.OutputFormat != "text" {
			t.Errorf("expected default OutputFormat=text, got %s", cfg.OutputFormat)
		}
	})

	t.Run("ReturnsErrorForPermissionDenied", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ".gomarklint.json")
		if err := os.WriteFile(configPath, []byte(`{"default":true}`), 0644); err != nil {
			t.Fatalf("failed to create test config: %v", err)
		}
		if err := os.Chmod(tmpDir, 0000); err != nil {
			t.Fatalf("failed to change permissions: %v", err)
		}
		defer func() {
			if err := os.Chmod(tmpDir, 0755); err != nil {
				t.Logf("failed to restore permissions: %v", err)
			}
		}()

		_, err := LoadOrDefault(configPath)
		if err == nil {
			t.Error("expected error for permission denied, got nil")
		}
	})
}

func TestMergeFlags(t *testing.T) {
	t.Run("MergesOutputFlag", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("output", "text", "")
		cmd.Flags().String("severity", "warning", "")
		_ = cmd.Flags().Set("output", "json")

		cfg := Default()
		flags := FlagValues{OutputFormat: "json", MinSeverity: "warning"}
		merged := MergeFlags(cfg, cmd, flags)

		if merged.OutputFormat != "json" {
			t.Errorf("expected OutputFormat=json, got %s", merged.OutputFormat)
		}
	})

	t.Run("MergesSeverityFlag", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("output", "text", "")
		cmd.Flags().String("severity", "warning", "")
		_ = cmd.Flags().Set("severity", "error")

		cfg := Default()
		flags := FlagValues{OutputFormat: "text", MinSeverity: "error"}
		merged := MergeFlags(cfg, cmd, flags)

		if merged.MinSeverity != SeverityError {
			t.Errorf("expected MinSeverity=error, got %s", merged.MinSeverity)
		}
	})

	t.Run("DoesNotMergeUnchangedFlags", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("output", "text", "")
		cmd.Flags().String("severity", "warning", "")

		cfg := Config{OutputFormat: "json", MinSeverity: SeverityError}
		flags := FlagValues{OutputFormat: "text", MinSeverity: "warning"}
		merged := MergeFlags(cfg, cmd, flags)

		if merged.OutputFormat != "json" {
			t.Errorf("expected OutputFormat unchanged at json, got %s", merged.OutputFormat)
		}
		if merged.MinSeverity != SeverityError {
			t.Errorf("expected MinSeverity unchanged at error, got %s", merged.MinSeverity)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("ValidTextFormat", func(t *testing.T) {
		cfg := Config{OutputFormat: "text", MinSeverity: SeverityWarning}
		if err := Validate(cfg); err != nil {
			t.Errorf("unexpected error for valid text format: %v", err)
		}
	})

	t.Run("ValidJSONFormat", func(t *testing.T) {
		cfg := Config{OutputFormat: "json", MinSeverity: SeverityError}
		if err := Validate(cfg); err != nil {
			t.Errorf("unexpected error for valid json format: %v", err)
		}
	})

	t.Run("InvalidOutputFormat", func(t *testing.T) {
		cfg := Config{OutputFormat: "xml", MinSeverity: SeverityWarning}
		if err := Validate(cfg); err == nil {
			t.Error("expected error for invalid output format")
		}
	})

	t.Run("InvalidSeverity", func(t *testing.T) {
		cfg := Config{OutputFormat: "text", MinSeverity: "verbose"}
		if err := Validate(cfg); err == nil {
			t.Error("expected error for invalid severity")
		}
	})
}
