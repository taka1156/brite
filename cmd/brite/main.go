package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/taka1156/brite/internal/app"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	app.NewApp().Run()
}
