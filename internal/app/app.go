package app

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/taka1156/brite/internal/command/add"
	"github.com/taka1156/brite/internal/command/convert"
	"github.com/taka1156/brite/internal/command/help"
	"github.com/taka1156/brite/internal/command/initialize"
	"github.com/taka1156/brite/internal/command/publish"
	"github.com/taka1156/brite/internal/command/setup"
	"github.com/taka1156/brite/internal/entity"
	infrastorage "github.com/taka1156/brite/internal/infra/storage"
)

type Command struct {
	HelpCommand
	InitializeCommand
	SetupCommand
	AddCommand
	ConvertCommand
	PublishCommand
}

type App struct {
	Cmd        Command
	storageErr error
}

func NewApp() *App {
	r2storage, storageErr := infrastorage.NewR2Storage()

	cmd := Command{
		HelpCommand:       help.NewHelpBrite(),
		InitializeCommand: initialize.NewInitializeConfig(),
		SetupCommand:      setup.NewSetupProject(),
		AddCommand:        add.NewAddArticle(),
		ConvertCommand:    convert.NewConvertArticle(),
	}
	if storageErr == nil {
		cmd.PublishCommand = publish.NewPublishArticle(r2storage)
	}

	return &App{
		Cmd:        cmd,
		storageErr: storageErr,
	}
}

func (a *App) Run() {
	jsonNames := entity.JsonNames{
		All:      entity.ALL_JSON_FILE_NAME,
		Category: entity.CATEGORY_JSON_FILE_NAME,
		Tag:      entity.TAG_JSON_FILE_NAME,
	}

	switch os.Args[1] {
	case "help":
		a.Cmd.Help()
		return
	case "init":
		absPath, err := parseConfigPath("init", os.Args[2:])
		if err != nil {
			fmt.Printf("Error resolving config path: %v\n", err)
			return
		}
		a.Cmd.Initialize(entity.ClientConfig{ConfigPath: absPath})
	case "setup":
		absPath, err := parseConfigPath("setup", os.Args[2:])
		if err != nil {
			fmt.Printf("Error resolving config path: %v\n", err)
			return
		}
		a.Cmd.Setup(entity.ClientConfig{ConfigPath: absPath})
	case "new":
		absPath, err := parseConfigPath("new", os.Args[2:])
		if err != nil {
			fmt.Printf("Error resolving config path: %v\n", err)
			return
		}
		a.Cmd.Add(entity.ClientConfig{ConfigPath: absPath})
	case "convert":
		absPath, err := parseConfigPath("convert", os.Args[2:])
		if err != nil {
			fmt.Printf("Error resolving config path: %v\n", err)
			return
		}
		a.Cmd.Convert(entity.ClientConfig{ConfigPath: absPath}, jsonNames)
	case "publish":
		if a.storageErr != nil {
			fmt.Printf("Error initializing storage: %v\n", a.storageErr)
			return
		}
		absPath, err := parseConfigPath("publish", os.Args[2:])
		if err != nil {
			fmt.Printf("Error resolving config path: %v\n", err)
			return
		}
		a.Cmd.Publish(entity.ClientConfig{ConfigPath: absPath})
	default:
		fmt.Println("Unknown command. Available commands: init, setup, new, convert, publish")
		return
	}
}

func parseConfigPath(cmdName string, args []string) (string, error) {
	fs := flag.NewFlagSet(cmdName, flag.ExitOnError)
	configPath := fs.String("config-path", entity.CONFIG_FILE_NAME, "path to brite.json")
	fs.Parse(args)
	return filepath.Abs(*configPath)
}
