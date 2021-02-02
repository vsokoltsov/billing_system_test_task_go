package app

import (
	"database/sql"
	"log"
	"net/http"

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
}

func (a *App) Run() {

}
