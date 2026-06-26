package main

import (
	"fmt"
	"os"

	"github.com/taka1156/cms-cli/internal"
	"github.com/taka1156/cms-cli/internal/entity"
)

func main() {

	jsonNames := entity.JsonNames{
		All:      entity.ALL_JSON_FILE_NAME,
		Category: entity.CATEGORY_JSON_FILE_NAME,
		Tag:      entity.TAG_JSON_FILE_NAME,
	}

	setupCommands := internal.NewSetupProjectCommand()
	addCommand := internal.NewAddArticleCommand()
	convertCommand := internal.NewConvertArticleCommand()

	// コマンド引数のチェック
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "setup":
			setupCommands.Setup()
			return
		case "new":
			addCommand.Add()
			return
		case "convert":
			convertCommand.Convert(jsonNames)
			return
		default:
			fmt.Println("Unknown command. Available commands: setup, new, convert")
			return
		}
	} else {
		fmt.Println("No command provided. Available commands: setup, new, convert")
	}

}
