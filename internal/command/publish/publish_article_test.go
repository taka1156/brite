package publish

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/taka1156/brite/internal/entity"
)

// mockStorage records calls for assertion in tests.
type mockStorage struct {
	uploadedKeys []string
	deletedKeys  []string
}

func (m *mockStorage) Upload(_ context.Context, _, key string, _ io.Reader, _ string) error {
	m.uploadedKeys = append(m.uploadedKeys, key)
	return nil
}

func (m *mockStorage) Delete(_ context.Context, _, key string) error {
	m.deletedKeys = append(m.deletedKeys, key)
	return nil
}

func TestContentType(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: "image.jpg", want: "image/jpeg"},
		{path: "image.jpeg", want: "image/jpeg"},
		{path: "IMAGE.JPG", want: "image/jpeg"},
		{path: "image.png", want: "image/png"},
		{path: "image.gif", want: "image/gif"},
		{path: "image.svg", want: "image/svg+xml"},
		{path: "image.webp", want: "image/webp"},
		{path: "data.json", want: "application/json"},
		{path: "file.bin", want: "application/octet-stream"},
		{path: "noextension", want: "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := contentType(tt.path)
			if got != tt.want {
				t.Errorf("contentType(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestDetectDiff(t *testing.T) {
	writeFile := func(t *testing.T, path, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string)
		caches    func(dir string) map[string]entity.ImageCache
		wantTypes map[ChangeType]int // expected count per ChangeType
	}{
		{
			name: "new file is Added",
			setup: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "new.png"), "data")
			},
			caches: func(dir string) map[string]entity.ImageCache {
				return map[string]entity.ImageCache{}
			},
			wantTypes: map[ChangeType]int{Added: 1},
		},
		{
			name: "file with same size is NoChange",
			setup: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "same.png"), "data")
			},
			caches: func(dir string) map[string]entity.ImageCache {
				path := filepath.Join(dir, "same.png")
				return map[string]entity.ImageCache{
					path: {FilePath: path, Size: 4},
				}
			},
			wantTypes: map[ChangeType]int{NoChange: 1},
		},
		{
			name: "file with different size is Modified",
			setup: func(t *testing.T, dir string) {
				writeFile(t, filepath.Join(dir, "mod.png"), "updated-data")
			},
			caches: func(dir string) map[string]entity.ImageCache {
				path := filepath.Join(dir, "mod.png")
				return map[string]entity.ImageCache{
					path: {FilePath: path, Size: 1},
				}
			},
			wantTypes: map[ChangeType]int{Modified: 1},
		},
		{
			name:  "cached file missing on disk is Deleted",
			setup: func(t *testing.T, dir string) {},
			caches: func(dir string) map[string]entity.ImageCache {
				path := filepath.Join(dir, "gone.png")
				return map[string]entity.ImageCache{
					path: {FilePath: path, Size: 10},
				}
			},
			wantTypes: map[ChangeType]int{Deleted: 1},
		},
		{
			name:  "empty directory produces no diffs",
			setup: func(t *testing.T, dir string) {},
			caches: func(dir string) map[string]entity.ImageCache {
				return map[string]entity.ImageCache{}
			},
			wantTypes: map[ChangeType]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)

			diffs, err := detectDiff(dir, tt.caches(dir))
			if err != nil {
				t.Fatalf("detectDiff() error = %v", err)
			}

			got := map[ChangeType]int{}
			for _, d := range diffs {
				got[d.ChangeType]++
			}

			// Compare sorted keys for stable output
			wantKeys := sortedChangeTypes(tt.wantTypes)
			gotKeys := sortedChangeTypes(got)
			if len(wantKeys) != len(gotKeys) {
				t.Errorf("detectDiff() = %v, want %v", got, tt.wantTypes)
				return
			}
			for _, k := range wantKeys {
				if got[k] != tt.wantTypes[k] {
					t.Errorf("ChangeType %v: count = %d, want %d", k, got[k], tt.wantTypes[k])
				}
			}
		})
	}
}

func sortedChangeTypes(m map[ChangeType]int) []ChangeType {
	keys := make([]ChangeType, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
