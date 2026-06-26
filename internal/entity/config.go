package entity

type JsonNames struct {
	All      string
	Category string
	Tag      string
}

// categories/tagsの1要素（名前+紐づく画像パス）
type TaxonomyDefinition struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

// 設定ファイルの構造
type CMSConfig struct {
	Schema     string               `json:"$schema"`
	ContentDir string               `json:"content_dir"`
	OutputDir  string               `json:"output_dir"`
	Categories []TaxonomyDefinition `json:"categories"`
	Tags       []TaxonomyDefinition `json:"tags"`
}

// 各記事のデータ構造
type Post struct {
	Slug      string   `json:"slug"`
	Title     string   `json:"title"`
	Category  string   `json:"category"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at" yaml:"created_at"`
	UpdatedAt string   `json:"updated_at" yaml:"updated_at"`
	Content   string   `json:"content" yaml:"-"`
}

// 最終出力のデータ構造（byCategory/byTagはslug参照のみで本文の重複を避ける）
type ResponseData struct {
	All        []Post              `json:"all"`
	ByCategory map[string][]string `json:"byCategory"`
	ByTag      map[string][]string `json:"byTag"`
}

// category.json / tag.json の1エントリ（画像情報 + 紐づく記事slug一覧）
type TaxonomyEntry struct {
	Image string   `json:"image"`
	Slugs []string `json:"slugs"`
}
