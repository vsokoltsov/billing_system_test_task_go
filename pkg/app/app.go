package app

import (
	"billing_system_test_task/pkg/api"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type App struct {
	host          string
	port          string
	env           string
	pathDelimiter string
	router        *mux.Router
	server        *http.Server
}

func (a *App) Initialize(env, host, port, pathDelimiter, dbProvider, sqlDbConnStr string) {
	var (
		sqlDB        *sql.DB
		sqlDbOpenErr error
	)

	a.env = env
	a.host = host
	a.port = port
	a.pathDelimiter = pathDelimiter

	sqlDB, sqlDbOpenErr = sql.Open(dbProvider, sqlDbConnStr)
	if sqlDbOpenErr != nil {
		log.Fatalf("Error sql database open: %s", sqlDbOpenErr)
		return
	}
	if pingErr := sqlDB.Ping(); pingErr != nil {
		log.Fatalf("Error sql database connection: %s", pingErr)
	}

	a.router = api.SetUpRoutes(pathDelimiter, sqlDB)

	url := strings.Join([]string{a.host, a.port}, ":")
	a.server = &http.Server{
		Handler:      handlers.LoggingHandler(os.Stdout, a.router),
		Addr:         url,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}

func (a *App) Run() {
	log.Printf("Starting web server on port %s...", a.Port)
	log.Fatal(a.server.ListenAndServe())
}
