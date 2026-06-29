package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/taka1156/brite/internal"
	"github.com/taka1156/brite/internal/entity"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	jsonNames := entity.JsonNames{
		All:      entity.ALL_JSON_FILE_NAME,
		Category: entity.CATEGORY_JSON_FILE_NAME,
		Tag:      entity.TAG_JSON_FILE_NAME,
	}

	cmd := struct {
		*internal.HelpBriteCommand
		*internal.InitializeConfigCommand
		*internal.SetupProjectCommand
		*internal.AddArticleCommand
		*internal.ConvertArticleCommand
		*internal.PublishArticleCommand
	}{
		internal.NewHelpBriteCommand(),
		internal.NewInitializeConfigCommand(),
		internal.NewSetupProjectCommand(),
		internal.NewAddArticleCommand(),
		internal.NewConvertArticleCommand(),
		internal.NewPublishArticleCommand(),
	}

	flag.Parse()
	clientConfig := entity.ClientConfig{}
	clientConfig.ConfigPath = flag.String("config-path", entity.CONFIG_FILE_NAME, "local path URL to base image config JSON")

	// コマンド引数のチェック
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "help":
			cmd.Help()
			return
		case "init":
			cmd.Initialize(clientConfig)
			return
		case "setup":
			cmd.Setup(clientConfig)
			return
		case "new":
			cmd.Add(clientConfig)
			return
		case "convert":
			cmd.Convert(clientConfig, jsonNames)
			return
		case "publish":
			cmd.Publish(clientConfig)
			return
		default:
			fmt.Println("Unknown command. Available commands: init, setup, new, convert, publish")
			return
		}
	} else {
		fmt.Println("No command provided. Available commands: init, setup, new, convert, publish")
	}

}
