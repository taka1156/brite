package entity

type ClientConfig struct {
	ConfigPath string
}

type JsonNames struct {
	All      string
	Category string
	Tag      string
}

type R2Config struct {
	Endpoint   string `json:"endpoint"`
	BucketName string `json:"bucketName"`
	BaseUrl    string `json:"baseUrl"`
}

type BadgeConfig struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type BriteConfig struct {
	Schema     string        `json:"$schema"`
	ArticleDir string        `json:"articleDir"`
	ImageDir   string        `json:"imageDir"`
	OutputDir  string        `json:"outputDir"`
	CacheDir   string        `json:"cacheDir"`
	Categories []BadgeConfig `json:"categories"`
	Tags       []BadgeConfig `json:"tags"`
	R2         R2Config      `json:"r2"`
}

// FrontMatterSummary represents the front matter of a Markdown file.
type FrontMatterSummary struct {
	Slug        string   `yaml:"-"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Thumbnail   string   `yaml:"thumbnail"`
	Category    string   `yaml:"category"`
	Tags        []string `yaml:"tags"`
	CreatedAt   string   `yaml:"created_at"`
	UpdatedAt   string   `yaml:"updated_at"`
}

// PostSummary represents the summary of a post, including its metadata.
type PostSummary struct {
	Slug        string        `json:"slug"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Thumbnail   string        `json:"thumbnail"`
	Category    BadgeConfig   `json:"category"`
	Tags        []BadgeConfig `json:"tags"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
}

// Post represents a complete post, including its summary and content.
type InputArticle struct {
	Summary FrontMatterSummary `yaml:"summary"`
	Content string             `yaml:"-"`
}

type PostArticle struct {
	Summary PostSummary `json:"summary"`
	Content string      `json:"content"`
}

type Badge struct {
	Name  string        `json:"name"`
	Image string        `json:"image"`
	Posts []PostSummary `json:"posts"`
}

type ResponseData struct {
	All        []PostArticle `json:"all"`
	ByCategory []Badge       `json:"byCategory"`
	ByTag      []Badge       `json:"byTag"`
}

type ImageCache struct {
	FilePath string `json:"filePath"`
	Size     int64  `json:"size"`
}
