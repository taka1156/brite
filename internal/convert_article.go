package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/taka1156/cms-cli/internal/entity"
	"gopkg.in/yaml.v3"
)

type ConvertArticleCommand struct{}

func NewConvertArticleCommand() *ConvertArticleCommand {
	return &ConvertArticleCommand{}
}

// 記事変換（convert）コマンドの処理
func (c *ConvertArticleCommand) Convert(jsonNames entity.JsonNames) {

	// 1. cmsc.json の読み込み（通常のビルド処理）
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 2. 出力データの初期化
	data := entity.ResponseData{
		All:        []entity.Post{},
		ByCategory: make(map[string][]string),
		ByTag:      make(map[string][]string),
	}

	categoryNames := taxonomyNames(config.Categories)
	tagNames := taxonomyNames(config.Tags)

	categoryImages := make(map[string]string, len(config.Categories))
	for _, c := range config.Categories {
		categoryImages[c.Name] = c.Image
		data.ByCategory[c.Name] = []string{}
	}

	tagImages := make(map[string]string, len(config.Tags))
	for _, t := range config.Tags {
		tagImages[t.Name] = t.Image
		data.ByTag[t.Name] = []string{}
	}

	contains := func(list []string, item string) bool {
		for _, x := range list {
			if x == item {
				return true
			}
		}
		return false
	}

	// 3. Markdownディレクトリの巡回
	err = filepath.WalkDir(config.ContentDir, func(path string, d fs.DirEntry, err error) error {
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

		var post entity.Post
		if err := yaml.Unmarshal(parts[1], &post); err != nil {
			fmt.Printf("Warning: Failed to parse YAML (%s): %v\n", path, err)
			return nil
		}

		relPath, _ := filepath.Rel(config.ContentDir, path)
		post.Slug = strings.TrimSuffix(relPath, filepath.Ext(relPath))

		// 本文（フロントマター以降の部分）をそのままcontentとして保持
		post.Content = strings.TrimSpace(string(parts[2]))

		data.All = append(data.All, post)

		if post.Category != "" && contains(categoryNames, post.Category) {
			data.ByCategory[post.Category] = append(data.ByCategory[post.Category], post.Slug)
		} else if post.Category != "" {
			fmt.Printf("Notice: Skipped unregistered category -> %s (%s)\n", post.Category, path)
		}

		for _, tag := range post.Tags {
			if tag != "" && contains(tagNames, tag) {
				data.ByTag[tag] = append(data.ByTag[tag], post.Slug)
			} else if tag != "" {
				fmt.Printf("Notice: Skipped unregistered tag -> %s (%s)\n", tag, path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking paths: %v\n", err)
		return
	}

	// 3.5. 作成日(date)の降順で全カテゴリのソート
	sortPostsByDateDesc(data.All)

	slugToCreatedAt := make(map[string]string, len(data.All))
	for _, p := range data.All {
		slugToCreatedAt[p.Slug] = p.CreatedAt
	}

	for cat := range data.ByCategory {
		sortSlugsByDateDesc(data.ByCategory[cat], slugToCreatedAt)
	}
	for tag := range data.ByTag {
		sortSlugsByDateDesc(data.ByTag[tag], slugToCreatedAt)
	}

	// 4. JSONへの変換と書き出し（all.json / category.json / tag.json の3ファイルに分割）
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		fmt.Printf("Error creating output_dir: %v\n", err)
		return
	}

	if err := writeJSONFile(filepath.Join(config.OutputDir, jsonNames.All), data); err != nil {
		fmt.Printf("Error writing %s: %v\n", jsonNames.All, err)
		return
	}

	categoryOutput := buildTaxonomyOutput(data.ByCategory, categoryImages)
	if err := writeJSONFile(filepath.Join(config.OutputDir, jsonNames.Category), categoryOutput); err != nil {
		fmt.Printf("Error writing %s: %v\n", jsonNames.Category, err)
		return
	}

	tagOutput := buildTaxonomyOutput(data.ByTag, tagImages)
	if err := writeJSONFile(filepath.Join(config.OutputDir, jsonNames.Tag), tagOutput); err != nil {
		fmt.Printf("Error writing %s: %v\n", jsonNames.Tag, err)
		return
	}

	fmt.Printf("Success! Exported %s, %s, %s to %s\n", jsonNames.All, jsonNames.Category, jsonNames.Tag, config.OutputDir)
}

// {名前: [slug,...]} と {名前: image} を合成して、
// category.json / tag.json 用の {名前: {image, slugs}} 構造を組み立てる
func buildTaxonomyOutput(slugsByName map[string][]string, imagesByName map[string]string) map[string]entity.TaxonomyEntry {
	output := make(map[string]entity.TaxonomyEntry, len(slugsByName))
	for name, slugs := range slugsByName {
		output[name] = entity.TaxonomyEntry{
			Image: imagesByName[name],
			Slugs: slugs,
		}
	}
	return output
}

// 任意のデータをインデント付きJSONとしてファイルに書き出す共通処理
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

// slug配列を、対応するcreated_atを基準に降順（新しい記事が先頭）でソートする。
// slugToCreatedAtに存在しない/パース不能なslugは最も古い扱いとして末尾に回す。
func sortSlugsByDateDesc(slugs []string, slugToCreatedAt map[string]string) {
	sort.SliceStable(slugs, func(i, j int) bool {
		ti, errI := time.Parse(time.RFC3339, slugToCreatedAt[slugs[i]])
		tj, errJ := time.Parse(time.RFC3339, slugToCreatedAt[slugs[j]])

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

// dateを基準に降順（新しい記事が先頭）でソートする。
// パース不能なdateは最も古い扱いとして末尾に回す。
func sortPostsByDateDesc(posts []entity.Post) {
	sort.SliceStable(posts, func(i, j int) bool {
		ti, errI := time.Parse(time.RFC3339, posts[i].CreatedAt)
		tj, errJ := time.Parse(time.RFC3339, posts[j].CreatedAt)

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
