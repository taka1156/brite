package convert

import (
	"testing"
	"time"

	"github.com/taka1156/brite/internal/entity"
)

func TestReplaceImagePaths(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		baseUrl  string
		imageDir string
		want     string
	}{
		{
			name:     "markdown image in imageDir is replaced",
			content:  "![alt](./images/article/foo.png)",
			baseUrl:  "https://cdn.example.com",
			imageDir: "images",
			want:     "![alt](https://cdn.example.com/article/foo.png)",
		},
		{
			name:     "markdown image not in imageDir is unchanged",
			content:  "![alt](./other/foo.png)",
			baseUrl:  "https://cdn.example.com",
			imageDir: "images",
			want:     "![alt](./other/foo.png)",
		},
		{
			name:     "html img in imageDir is replaced",
			content:  `<img src="./images/article/bar.jpg" alt="test">`,
			baseUrl:  "https://cdn.example.com",
			imageDir: "images",
			want:     `<img  src="https://cdn.example.com/article/bar.jpg" alt="test">`,
		},
		{
			name:     "html img not in imageDir is unchanged",
			content:  `<img src="./other/bar.jpg">`,
			baseUrl:  "https://cdn.example.com",
			imageDir: "images",
			want:     `<img src="./other/bar.jpg">`,
		},
		{
			name:     "baseUrl trailing slash is trimmed",
			content:  "![alt](./images/article/foo.png)",
			baseUrl:  "https://cdn.example.com/",
			imageDir: "images",
			want:     "![alt](https://cdn.example.com/article/foo.png)",
		},
		{
			name:     "no images returns content unchanged",
			content:  "# Hello World\nsome text",
			baseUrl:  "https://cdn.example.com",
			imageDir: "images",
			want:     "# Hello World\nsome text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceImagePaths(tt.content, tt.baseUrl, tt.imageDir)
			if got != tt.want {
				t.Errorf("replaceImagePaths() =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

func TestSortSlugsByDateDesc(t *testing.T) {
	tests := []struct {
		name          string
		articles      []entity.PostSummary
		slugToCreated map[string]string
		wantOrder     []string
	}{
		{
			name: "sorts newest first",
			articles: []entity.PostSummary{
				{Slug: "old"},
				{Slug: "new"},
				{Slug: "mid"},
			},
			slugToCreated: map[string]string{
				"old": "2024-01-01T00:00:00Z",
				"new": "2024-03-01T00:00:00Z",
				"mid": "2024-02-01T00:00:00Z",
			},
			wantOrder: []string{"new", "mid", "old"},
		},
		{
			name: "invalid date goes to end",
			articles: []entity.PostSummary{
				{Slug: "invalid"},
				{Slug: "valid"},
			},
			slugToCreated: map[string]string{
				"invalid": "not-a-date",
				"valid":   "2024-01-01T00:00:00Z",
			},
			wantOrder: []string{"valid", "invalid"},
		},
		{
			name: "single item is unchanged",
			articles: []entity.PostSummary{
				{Slug: "only"},
			},
			slugToCreated: map[string]string{
				"only": "2024-01-01T00:00:00Z",
			},
			wantOrder: []string{"only"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			articles := make([]entity.PostSummary, len(tt.articles))
			copy(articles, tt.articles)

			sortSlugsByDateDesc(articles, tt.slugToCreated)

			for i, want := range tt.wantOrder {
				if articles[i].Slug != want {
					t.Errorf("position %d: got %q, want %q", i, articles[i].Slug, want)
				}
			}
		})
	}
}

func TestSortPostsByDateDesc(t *testing.T) {
	makePost := func(slug, date string) entity.Post {
		return entity.Post{Summary: entity.PostSummary{Slug: slug, CreatedAt: date}}
	}

	tests := []struct {
		name      string
		posts     []entity.Post
		wantOrder []string
	}{
		{
			name: "sorts newest first",
			posts: []entity.Post{
				makePost("a", "2024-01-01T00:00:00Z"),
				makePost("c", "2024-03-01T00:00:00Z"),
				makePost("b", "2024-02-01T00:00:00Z"),
			},
			wantOrder: []string{"c", "b", "a"},
		},
		{
			name: "invalid date goes to end",
			posts: []entity.Post{
				makePost("invalid", "not-a-date"),
				makePost("valid", "2024-01-01T00:00:00Z"),
			},
			wantOrder: []string{"valid", "invalid"},
		},
		{
			name: "same date preserves original order",
			posts: []entity.Post{
				makePost("first", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)),
				makePost("second", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)),
			},
			wantOrder: []string{"first", "second"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts := make([]entity.Post, len(tt.posts))
			copy(posts, tt.posts)

			sortPostsByDateDesc(posts)

			for i, want := range tt.wantOrder {
				if posts[i].Summary.Slug != want {
					t.Errorf("position %d: got %q, want %q", i, posts[i].Summary.Slug, want)
				}
			}
		})
	}
}
