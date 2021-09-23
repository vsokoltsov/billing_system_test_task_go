package main

import (
	"billing_system_test_task/pkg/app"
	"billing_system_test_task/pkg/entities"
	"log"
	"os"
	"strings"
)

func getEnv(key, def string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return def
	}
	return value
}

func getProjectPath(delimeter string) string {
	projectDirectory, directoryErr := os.Getwd()

	if directoryErr != nil {
		log.Fatalf("Could not locate current directory: %s", directoryErr)
	}

	isUnderCmd := strings.Contains(projectDirectory, delimeter)
	if isUnderCmd {
		var cmdIdx int
		splitPath := strings.Split(projectDirectory, "/")
		for idx, pathElem := range splitPath {
			if pathElem == delimeter {
				cmdIdx = idx
				break
			}
		}
		projectDirectory = strings.Join(splitPath[:cmdIdx], "/")
	}

	return projectDirectory
}

func main() {
	// pathDelimiter := getEnv("PATH_SEPARATOR", "cmd")
	// projectPath := getProjectPath(pathDelimiter)
	// projectDirectory, directoryErr := os.Getwd()

	// if directoryErr != nil {
	// 	log.Fatalf("Could not locate current directory: %s", directoryErr)
	// }

	// err := godotenv.Load(path.Join(projectPath, ".env"))
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	// var (
	// 	env           = getEnv("APP_ENV", "development")
	// 	host          = getEnv("APP_HOST", "app")
	// 	port          = getEnv("APP_PORT", "8000")
	// 	sqlDbUser     = getEnv("DB_USER", "user")
	// 	sqlDbPassword = getEnv("DB_PASSWORD", "password")
	// 	sqlDbHost     = getEnv("DB_HOST", "postgres")
	// 	sqlDbPort     = getEnv("DB_PORT", "3306")
	// 	sqlDbName     = getEnv("POSTGRES_DB", "billing")
	// 	dbProvider    = getEnv("DB_PROVIDER", "postgres")
	// )

	// dbConnString := fmt.Sprintf("port=%s host=%s user=%s "+
	// 	"password=%s dbname=%s sslmode=disable",
	// 	sqlDbPort, sqlDbHost, sqlDbUser, sqlDbPassword, sqlDbName)

	config := entities.NewEnvConfig()
	config.LoadEnvVariables("cmd")
	app := app.NewApp(config)
	app.Initialize()
	app.Run()
}
