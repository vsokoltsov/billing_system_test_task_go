package main

import (
	"billing_system_test_task/pkg/app"
	"billing_system_test_task/pkg/entities"
)

func main() {
	config := entities.NewEnvConfig()
	config.LoadEnvVariables("cmd")
	app := app.NewApp(config)
	app.Run()
}
