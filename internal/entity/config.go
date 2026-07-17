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

type PostSummary struct {
	Slug        string   `json:"slug" yaml:"-"`
	Title       string   `json:"title" yaml:"title"`
	Description string   `json:"description" yaml:"description"`
	Thumbnail   string   `json:"thumbnail" yaml:"thumbnail"`
	Category    string   `json:"category" yaml:"category"`
	Tags        []string `json:"tags" yaml:"tags"`
	CreatedAt   string   `json:"created_at" yaml:"created_at"`
	UpdatedAt   string   `json:"updated_at" yaml:"updated_at"`
}

type Post struct {
	Summary PostSummary `json:"summary" yaml:"summary"`
	Content string      `json:"content" yaml:"-"`
}

type Badge struct {
	Name  string        `json:"name"`
	Image string        `json:"image"`
	Posts []PostSummary `json:"posts"`
}

type ResponseData struct {
	All        []Post  `json:"all"`
	ByCategory []Badge `json:"byCategory"`
	ByTag      []Badge `json:"byTag"`
}

type ImageCache struct {
	FilePath string `json:"filePath"`
	Size     int64  `json:"size"`
}
