package publish

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/taka1156/brite/internal/entity"
	"github.com/taka1156/brite/internal/infra/storage"
	"github.com/taka1156/brite/internal/utils"
)

type ChangeType int

const (
	Added ChangeType = iota
	Modified
	NoChange
	Deleted
)

type ImageDiff struct {
	FilePath   string
	Size       int64
	ChangeType ChangeType
}

type PublishArticle struct {
	storage storage.Storage
}

func NewPublishArticle(storage storage.Storage) *PublishArticle {
	return &PublishArticle{storage: storage}
}

func (c *PublishArticle) Publish(clientConfig entity.ClientConfig) {
	briteConfig, err := utils.LoadJson[entity.BriteConfig](clientConfig.ConfigPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	caches := []entity.ImageCache{}
	cacheFilePath := filepath.Join(briteConfig.CacheDir, entity.CACHE_FILE_NAME)
	if _, err := os.Stat(cacheFilePath); os.IsNotExist(err) {
		if err := saveCache(cacheFilePath, caches); err != nil {
			fmt.Printf("Error creating .caches.json: %v\n", err)
			return
		}
	} else {
		caches, err = utils.LoadJson[[]entity.ImageCache](cacheFilePath)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}

	cacheByPath := make(map[string]entity.ImageCache)
	for _, cache := range caches {
		cacheByPath[cache.FilePath] = cache
	}

	diffs, err := detectDiff(briteConfig.ImageDir, cacheByPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	ctx := context.Background()

	if err := applyDiffs(ctx, c.storage, briteConfig.R2.BucketName, diffs); err != nil {
		fmt.Println("Error:", err)
		return
	}

	if err := postOutput(ctx, c.storage, briteConfig); err != nil {
		fmt.Println("Error:", err)
		return
	}

	newCaches := []entity.ImageCache{}
	for _, diff := range diffs {
		switch diff.ChangeType {
		case Added, Modified, NoChange:
			newCaches = append(newCaches, entity.ImageCache{
				FilePath: diff.FilePath,
				Size:     diff.Size,
			})
		case Deleted:
			// skip deleted images
		}
	}

	if err := saveCache(cacheFilePath, newCaches); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Successfully posted output files and images to R2.")
}

func contentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

func detectDiff(imageDir string, caches map[string]entity.ImageCache) ([]ImageDiff, error) {
	current := map[string]entity.ImageCache{}

	err := filepath.Walk(imageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		current[path] = entity.ImageCache{
			FilePath: path,
			Size:     info.Size(),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var diffs []ImageDiff

	for path, img := range current {
		if prev, ok := caches[path]; !ok {
			diffs = append(diffs, ImageDiff{FilePath: path, Size: img.Size, ChangeType: Added})
		} else if prev.Size != img.Size {
			diffs = append(diffs, ImageDiff{FilePath: path, Size: img.Size, ChangeType: Modified})
		} else {
			diffs = append(diffs, ImageDiff{FilePath: path, Size: img.Size, ChangeType: NoChange})
		}
	}

	for path, cache := range caches {
		if _, ok := current[path]; !ok {
			diffs = append(diffs, ImageDiff{FilePath: cache.FilePath, Size: 0, ChangeType: Deleted})
		}
	}

	return diffs, nil
}

func saveCache(path string, caches []entity.ImageCache) error {
	data, err := json.MarshalIndent(caches, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func applyDiffs(ctx context.Context, s storage.Storage, bucketName string, diffs []ImageDiff) error {
	for _, diff := range diffs {
		switch diff.ChangeType {
		case Added, Modified:
			f, err := os.Open(diff.FilePath)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", diff.FilePath, err)
			}
			err = s.Upload(ctx, bucketName, diff.FilePath, f, contentType(diff.FilePath))
			f.Close()
			if err != nil {
				return err
			}
		case Deleted:
			if err := s.Delete(ctx, bucketName, diff.FilePath); err != nil {
				return err
			}
		}
	}
	return nil
}

func postOutput(ctx context.Context, s storage.Storage, briteConfig entity.BriteConfig) error {
	jsonFiles := []string{
		entity.ALL_JSON_FILE_NAME,
		entity.CATEGORY_JSON_FILE_NAME,
		entity.TAG_JSON_FILE_NAME,
	}
	for _, jsonFile := range jsonFiles {
		filePath := filepath.Join(briteConfig.OutputDir, jsonFile)
		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", filePath, err)
		}
		err = s.Upload(ctx, briteConfig.R2.BucketName, filePath, f, contentType(filePath))
		f.Close()
		if err != nil {
			return fmt.Errorf("failed to upload %s: %w", filePath, err)
		}
	}
	fmt.Println("Successfully uploaded output files.")
	return nil
}
