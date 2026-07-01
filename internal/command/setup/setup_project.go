package setup

import (
	"fmt"
	"os"

	"github.com/taka1156/brite/internal/entity"
	"github.com/taka1156/brite/internal/utils"
)

type SetupProject struct{}

func NewSetupProject() *SetupProject {
	return &SetupProject{}
}

func (c *SetupProject) Setup(clientConfig entity.ClientConfig) {
	config, err := utils.LoadJson[entity.BriteConfig](clientConfig.ConfigPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	dirs := []string{
		config.ArticleDir,
		config.ImageDir + "/article",
		config.ImageDir + "/category",
		config.ImageDir + "/tag",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			return
		}
	}

	fmt.Println("Success! Project setup completed.")
}
