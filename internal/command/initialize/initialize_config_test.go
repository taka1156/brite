package initialize

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/taka1156/brite/internal/entity"
)

func TestInitializeConfig_Initialize(t *testing.T) {
	const existingContent = `{"existing": true}`

	tests := []struct {
		name        string
		preExisting bool
		wantDefault bool
	}{
		{
			name:        "config does not exist creates file with defaults",
			preExisting: false,
			wantDefault: true,
		},
		{
			name:        "config already exists does not overwrite",
			preExisting: true,
			wantDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			configPath := filepath.Join(dir, "brite.json")

			if tt.preExisting {
				if err := os.WriteFile(configPath, []byte(existingContent), 0644); err != nil {
					t.Fatal(err)
				}
			}

			cmd := NewInitializeConfig()
			cmd.Initialize(entity.ClientConfig{ConfigPath: configPath})

			data, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("config file should exist: %v", err)
			}

			if tt.wantDefault {
				var cfg entity.BriteConfig
				if err := json.Unmarshal(data, &cfg); err != nil {
					t.Fatalf("expected valid BriteConfig JSON: %v", err)
				}
				if cfg.ArticleDir == "" {
					t.Error("expected non-empty articleDir in created config")
				}
				if len(cfg.Categories) == 0 {
					t.Error("expected default categories in created config")
				}
			} else {
				if string(data) != existingContent {
					t.Error("existing config should not be overwritten")
				}
			}
		})
	}
}
