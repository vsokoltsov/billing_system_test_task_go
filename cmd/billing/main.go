package main

import (
	"billing_system_test_task/internal/app"
	"billing_system_test_task/internal/entities"
	"log"
)

func main() {
	config := entities.NewEnvConfig()
	loadEnvErr := config.LoadEnvVariables("cmd")
	if loadEnvErr != nil {
		log.Fatal("Unable to variables from .env file")
	}
	app := app.NewApp(config)
	app.Run()
}
