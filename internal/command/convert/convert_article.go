package convert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/taka1156/brite/internal/entity"
	"github.com/taka1156/brite/internal/utils"
	"gopkg.in/yaml.v3"
)

type ConvertArticle struct{}

func NewConvertArticle() *ConvertArticle {
	return &ConvertArticle{}
}

func (c *ConvertArticle) Convert(clientConfig entity.ClientConfig, jsonNames entity.JsonNames) {

	config, err := utils.LoadJson[entity.BriteConfig](clientConfig.ConfigPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Initialize the output data
	data := &entity.ResponseData{
		All:        []entity.PostArticle{},
		ByCategory: []entity.Badge{},
		ByTag:      []entity.Badge{},
	}

	for _, category := range config.Categories {
		data.ByCategory = append(data.ByCategory, entity.Badge{
			Name:  category.Name,
			Image: category.Image,
			Posts: []entity.PostSummary{},
		})
	}

	for _, tag := range config.Tags {
		data.ByTag = append(data.ByTag, entity.Badge{
			Name:  tag.Name,
			Image: tag.Image,
			Posts: []entity.PostSummary{},
		})
	}

	// Recursively scan the article_dir for Markdown files and read their content
	data, err = walkMarkdownFiles(config.ArticleDir, data, config, config.Categories, config.Tags)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Created a map of slug to created_at to keep track of the reference date.
	slugToCreatedAt := make(map[string]string, len(data.All))
	for _, p := range data.All {
		slugToCreatedAt[p.Summary.Slug] = p.Summary.CreatedAt
	}

	for i := range data.ByCategory {
		sortSlugsByDateDesc(data.ByCategory[i].Posts, slugToCreatedAt)
	}
	for i := range data.ByTag {
		sortSlugsByDateDesc(data.ByTag[i].Posts, slugToCreatedAt)
	}

	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		fmt.Printf("Error creating output_dir: %v\n", err)
		return
	}

	if err := writeJSONFile(filepath.Join(config.OutputDir, jsonNames.All), data); err != nil {
		fmt.Printf("Error writing %s: %v\n", jsonNames.All, err)
		return
	}

	if err := writeJSONFile(filepath.Join(config.OutputDir, jsonNames.Category), data.ByCategory); err != nil {
		fmt.Printf("Error writing %s: %v\n", jsonNames.Category, err)
		return
	}

	if err := writeJSONFile(filepath.Join(config.OutputDir, jsonNames.Tag), data.ByTag); err != nil {
		fmt.Printf("Error writing %s: %v\n", jsonNames.Tag, err)
		return
	}

	fmt.Printf("Success! Exported %s, %s, %s to %s\n", jsonNames.All, jsonNames.Category, jsonNames.Tag, config.OutputDir)
}

func replaceImagePaths(content, baseUrl, imageDir string) string {
	var imgTagRegex = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+\.(png|svg|jpe?g|gif|webp))\)`)
	var htmlImgRegex = regexp.MustCompile(`<img([^>]*?)src="([^"]+\.(png|svg|jpe?g|gif|webp))"([^>]*?)>`)

	// convert Markdown image syntax to use the baseUrl for images in imageDir
	content = imgTagRegex.ReplaceAllStringFunc(content, func(match string) string {
		sub := imgTagRegex.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		alt := sub[1]
		path := sub[2]

		// Normalize the relative path and check if it is under imageDir
		clean := filepath.Clean(path)
		if !strings.Contains(clean, imageDir) {
			return match
		}

		// Pick the part of the path after imageDir and construct the new URL
		idx := strings.Index(clean, imageDir)
		rel := clean[idx+len(imageDir):]
		url := strings.TrimSuffix(baseUrl, "/") + "/" + strings.TrimPrefix(rel, "/")

		return fmt.Sprintf("![%s](%s)", alt, url)
	})

	// convert HTML <img> tags to use the baseUrl for images in imageDir
	content = htmlImgRegex.ReplaceAllStringFunc(content, func(match string) string {
		sub := htmlImgRegex.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		before := sub[1]
		path := sub[2]
		after := sub[4]

		// Normalize the relative path and check if it is under imageDir
		clean := filepath.Clean(path)
		if !strings.Contains(clean, imageDir) {
			return match
		}

		// Pick the part of the path after imageDir and construct the new URL
		idx := strings.Index(clean, imageDir)
		rel := clean[idx+len(imageDir):]
		url := strings.TrimSuffix(baseUrl, "/") + "/" + strings.TrimPrefix(rel, "/")

		return fmt.Sprintf("<img%s src=\"%s\"%s>", before, url, after)
	})

	return content
}

// walkMarkdownFiles scans all articles, reads their summary and content, and creates lists of articles by category/tag.
func walkMarkdownFiles(contentDir string, data *entity.ResponseData, config entity.BriteConfig, categories, tags []entity.BadgeConfig) (*entity.ResponseData, error) {
	contains := func(list []entity.BadgeConfig, item string) bool {
		for _, x := range list {
			if x.Name == item {
				return true
			}
		}
		return false
	}

	err := filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		parts := bytes.SplitN(content, []byte("---\n"), 3)
		if len(parts) < 3 {
			parts = bytes.SplitN(content, []byte("---\r\n"), 3)
			if len(parts) < 3 {
				return nil
			}
		}

		var post entity.InputArticle
		if err := yaml.Unmarshal(parts[1], &post.Summary); err != nil {
			fmt.Printf("Warning: Failed to parse YAML (%s): %v\n", path, err)
			return nil
		}

		relPath, _ := filepath.Rel(contentDir, path)
		post.Summary.Slug = strings.TrimSuffix(relPath, filepath.Ext(relPath))

		post.Content = strings.TrimSpace(string(parts[2]))
		post.Content = replaceImagePaths(post.Content, config.R2.BaseUrl, config.ImageDir)

		postArticle := entity.PostArticle{
			Summary: entity.PostSummary{
				Slug:        post.Summary.Slug,
				Title:       post.Summary.Title,
				Description: post.Summary.Description,
				Thumbnail:   post.Summary.Thumbnail,
				Category:    findBadgeConfig(config.Categories, post.Summary.Category),
				Tags:        convertToBadgeConfigs(post.Summary.Tags, config.Tags),
				CreatedAt:   post.Summary.CreatedAt,
				UpdatedAt:   post.Summary.UpdatedAt,
			},
			Content: post.Content,
		}
		data.All = append(data.All, postArticle)

		// Categorize the post by its category
		if postArticle.Summary.Category.Name != "" && contains(categories, postArticle.Summary.Category.Name) {
			for i, category := range config.Categories {
				if category.Name == postArticle.Summary.Category.Name {
					data.ByCategory[i].Posts = append(data.ByCategory[i].Posts, postArticle.Summary)
					break
				}
			}
		} else if postArticle.Summary.Category.Name != "" {
			fmt.Printf("Notice: Skipped unregistered category -> %s (%s)\n", postArticle.Summary.Category.Name, path)
		}

		// Tags posts by their tags
		for _, tag := range postArticle.Summary.Tags {
			if tag.Name != "" && contains(tags, tag.Name) {
				for i, tagConfig := range config.Tags {
					if tagConfig.Name == tag.Name {
						data.ByTag[i].Posts = append(data.ByTag[i].Posts, postArticle.Summary)
						break
					}
				}
			} else if tag.Name != "" {
				fmt.Printf("Notice: Skipped unregistered tag -> %s (%s)\n", tag.Name, path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking paths: %v\n", err)
		return nil, err
	}

	sortPostsByDateDesc(data.All)

	return data, nil
}

// category or tag image retrieval
func findBadgeConfig(badges []entity.BadgeConfig, name string) entity.BadgeConfig {
	for _, badge := range badges {
		if badge.Name == name {
			return badge
		}
	}
	return entity.BadgeConfig{Name: name}
}

func convertToBadgeConfigs(names []string, tags []entity.BadgeConfig) []entity.BadgeConfig {
	badges := make([]entity.BadgeConfig, len(names))

	for i, name := range names {
		badges[i] = findBadgeConfig(tags, name)
		if badges[i].Name == "" {
			badges[i] = entity.BadgeConfig{Name: name}
		}
	}
	return badges
}

func writeJSONFile(path string, v interface{}) error {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to convert to JSON: %w", err)
	}

	if err := os.WriteFile(path, jsonBytes, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Sort slugs by date in descending order (newest articles first).
// Slugs that do not exist in slugToCreatedAt or cannot be parsed are treated as the oldest and moved to the end.
func sortSlugsByDateDesc(articles []entity.PostSummary, slugToCreatedAt map[string]string) {
	sort.SliceStable(articles, func(i, j int) bool {
		ti, errI := time.Parse(time.RFC3339, slugToCreatedAt[articles[i].Slug])
		tj, errJ := time.Parse(time.RFC3339, slugToCreatedAt[articles[j].Slug])

		if errI != nil && errJ != nil {
			return false
		}
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}

		return ti.After(tj)
	})
}

// Sort by date in descending order (newest articles first).
// Dates that cannot be parsed are treated as the oldest and moved to the end.
func sortPostsByDateDesc(posts []entity.PostArticle) {
	sort.SliceStable(posts, func(i, j int) bool {
		ti, errI := time.Parse(time.RFC3339, posts[i].Summary.CreatedAt)
		tj, errJ := time.Parse(time.RFC3339, posts[j].Summary.CreatedAt)

		if errI != nil && errJ != nil {
			return false
		}
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}

		return ti.After(tj)
	})
}
