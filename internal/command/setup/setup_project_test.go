package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/taka1156/brite/internal/entity"
)

func writeBriteConfig(t *testing.T, dir string, cfg entity.BriteConfig) string {
	t.Helper()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(dir, "brite.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	return configPath
}

func TestSetupProject_Setup(t *testing.T) {
	tests := []struct {
		name     string
		wantDirs []string
	}{
		{
			name: "creates article and image subdirectories",
			wantDirs: []string{
				"articles",
				"images/article",
				"images/category",
				"images/tag",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			cfg := entity.BriteConfig{
				ArticleDir: filepath.Join(dir, "articles"),
				ImageDir:   filepath.Join(dir, "images"),
			}
			configPath := writeBriteConfig(t, dir, cfg)

			cmd := NewSetupProject()
			cmd.Setup(entity.ClientConfig{ConfigPath: configPath})

			for _, d := range tt.wantDirs {
				full := filepath.Join(dir, d)
				if _, err := os.Stat(full); os.IsNotExist(err) {
					t.Errorf("expected directory %q to exist", d)
				}
			}
		})
	}
}
