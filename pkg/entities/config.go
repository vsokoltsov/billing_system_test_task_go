package entities

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type ConfigAdapter interface {
	LoadEnvVariables(appDelimiter string) error
	GetPathDelimiter() string
	GetDBConnectionString() string
	GetWaitTime() time.Duration
	GetAppHost() string
	GetAppPort() string
	GetDBProvider() string
}

type EnvConfig struct {
	pathDelimiter string
}

func NewEnvConfig() ConfigAdapter {
	return EnvConfig{}
}

func (ec EnvConfig) GetPathDelimiter() string {
	return getEnv("PATH_SEPARATOR", "cmd")
}

func (ec EnvConfig) GetDBConnectionString() string {
	var (
		sqlDbUser     = getEnv("DB_USER", "user")
		sqlDbPassword = getEnv("DB_PASSWORD", "password")
		sqlDbHost     = getEnv("DB_HOST", "postgres")
		sqlDbPort     = getEnv("DB_PORT", "3306")
		sqlDbName     = getEnv("POSTGRES_DB", "billing")
	)

	return fmt.Sprintf("port=%s host=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		sqlDbPort, sqlDbHost, sqlDbUser, sqlDbPassword, sqlDbName)

}

func (ec EnvConfig) GetWaitTime() time.Duration {
	return time.Second * 5
}

func (ec EnvConfig) GetAppHost() string {
	return getEnv("APP_HOST", "app")
}

func (ec EnvConfig) GetAppPort() string {
	return getEnv("APP_PORT", "8000")
}

func (ec EnvConfig) GetDBProvider() string {
	return getEnv("DB_PROVIDER", "postgres")
}

func (ec EnvConfig) LoadEnvVariables(appDelimiter string) error {
	projectPath := ec.getProjectPath(appDelimiter)
	err := godotenv.Load(path.Join(projectPath, ".env"))
	if err != nil {
		return err
	}
	return nil
}

func (ec EnvConfig) getProjectPath(defaultDelimiter string) string {
	projectDirectory, directoryErr := os.Getwd()

	if directoryErr != nil {
		log.Fatalf("Could not locate current directory: %s", directoryErr)
	}

	isUnderCmd := strings.Contains(projectDirectory, defaultDelimiter)
	if isUnderCmd {
		var cmdIdx int
		splitPath := strings.Split(projectDirectory, "/")
		for idx, pathElem := range splitPath {
			if pathElem == defaultDelimiter {
				cmdIdx = idx
				break
			}
		}
		projectDirectory = strings.Join(splitPath[:cmdIdx], "/")
	}

	return projectDirectory
}

func getEnv(key, def string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return def
	}
	return value
}
