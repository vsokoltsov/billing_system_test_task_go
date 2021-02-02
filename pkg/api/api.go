package api

import (
	"database/sql"

	"github.com/gorilla/mux"
)

func SetUpRoutes(pathDelimiter string, sqlDb *sql.DB) *mux.Router {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	return api
}
