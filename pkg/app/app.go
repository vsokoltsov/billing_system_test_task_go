package app

import (
	"billing_system_test_task/pkg/api"
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// App represents base application info
type App struct {
	host          string
	port          string
	env           string
	pathDelimiter string
	router        *mux.Router
	server        *http.Server
	wait          time.Duration
}

// App populates struct parameters with data
func (a *App) Initialize(env, host, port, pathDelimiter, dbProvider, sqlDbConnStr string) {
	var (
		sqlDB        *sql.DB
		sqlDbOpenErr error
	)

	a.env = env
	a.host = host
	a.port = port
	a.pathDelimiter = pathDelimiter
	a.wait = time.Second * 5

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

// Run starts application (with gracefull shutdown)
func (a *App) Run() {
	log.Printf("Starting web server on port %s...", a.port)
	go func() {
		if err := a.server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), a.wait)
	defer cancel()

	shutdownErr := a.server.Shutdown(ctx)
	if shutdownErr != nil {
		log.Fatalf("Error of server shutdown: %s", shutdownErr)
		os.Exit(0)
	}

	log.Println("Shutting down the service...")
	os.Exit(0)
}
