package add

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/taka1156/brite/internal/entity"
)

func TestPromptSingleSelect(t *testing.T) {
	tests := []struct {
		name    string
		options []string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "empty options returns empty string",
			options: []string{},
			input:   "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "valid selection returns correct option",
			options: []string{"tech", "life"},
			input:   "1\n",
			want:    "tech",
			wantErr: false,
		},
		{
			name:    "second option selected",
			options: []string{"tech", "life"},
			input:   "2\n",
			want:    "life",
			wantErr: false,
		},
		{
			name:    "empty input skips selection",
			options: []string{"tech", "life"},
			input:   "\n",
			want:    "",
			wantErr: false,
		},
		{
			name:    "out-of-range index returns error",
			options: []string{"tech", "life"},
			input:   "5\n",
			want:    "",
			wantErr: true,
		},
		{
			name:    "non-numeric input returns error",
			options: []string{"tech", "life"},
			input:   "abc\n",
			want:    "",
			wantErr: true,
		},
		{
			name:    "zero index returns error",
			options: []string{"tech", "life"},
			input:   "0\n",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			got, err := promptSingleSelect(reader, "Category", tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("promptSingleSelect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("promptSingleSelect() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPromptMultiSelect(t *testing.T) {
	tests := []struct {
		name    string
		options []string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:    "empty options returns empty slice",
			options: []string{},
			input:   "",
			want:    []string{},
			wantErr: false,
		},
		{
			name:    "single valid selection",
			options: []string{"Go", "Rust"},
			input:   "1\n",
			want:    []string{"Go"},
			wantErr: false,
		},
		{
			name:    "multiple valid selections",
			options: []string{"Go", "Rust", "TypeScript"},
			input:   "1,3\n",
			want:    []string{"Go", "TypeScript"},
			wantErr: false,
		},
		{
			name:    "duplicate selections are deduplicated",
			options: []string{"Go", "Rust"},
			input:   "1,1\n",
			want:    []string{"Go"},
			wantErr: false,
		},
		{
			name:    "empty input skips selection",
			options: []string{"Go", "Rust"},
			input:   "\n",
			want:    []string{},
			wantErr: false,
		},
		{
			name:    "out-of-range index returns error",
			options: []string{"Go", "Rust"},
			input:   "99\n",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "non-numeric input returns error",
			options: []string{"Go", "Rust"},
			input:   "abc\n",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			got, err := promptMultiSelect(reader, "Tags", tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("promptMultiSelect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("promptMultiSelect() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("promptMultiSelect()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestWritePostFile(t *testing.T) {
	basePost := entity.PostSummary{
		Slug:      "test-slug",
		Title:     "Test Article",
		Category:  "tech",
		Tags:      []string{"Go"},
		CreatedAt: "2024-01-01T00:00:00Z",
	}

	tests := []struct {
		name    string
		setup   func(t *testing.T, dir string)
		wantErr bool
	}{
		{
			name:    "creates file with front matter",
			setup:   func(t *testing.T, dir string) {},
			wantErr: false,
		},
		{
			name: "returns error if file already exists",
			setup: func(t *testing.T, dir string) {
				path := filepath.Join(dir, basePost.Slug+".md")
				if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			err := writePostFile(dir, basePost)
			if (err != nil) != tt.wantErr {
				t.Errorf("writePostFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				data, err := os.ReadFile(filepath.Join(dir, basePost.Slug+".md"))
				if err != nil {
					t.Fatalf("expected file to exist: %v", err)
				}
				content := string(data)
				if !strings.Contains(content, "title: Test Article") {
					t.Errorf("expected front matter to contain title, got:\n%s", content)
				}
				if !strings.Contains(content, "---") {
					t.Errorf("expected front matter delimiters, got:\n%s", content)
				}
			}
		})
	}
}
