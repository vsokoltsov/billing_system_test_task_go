package http

import (
	"net/http"

	_ "billing_system_test_task/docs" // docs is generated by Swag CLI, you have to import it.

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Billing System API
// @version 1.0
// @description Simple billing system
// @termsOfService http://swagger.io/terms/
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8000
// @BasePath /
func NewRouter(usersHandler UsersHandler, walletsHandler WalletsHandler) http.Handler {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/users/", usersHandler.Create).Methods("POST").Name("CREATE_USER")
	api.HandleFunc("/users/{id}/enroll/", usersHandler.Enroll).Methods("POST").Name("ENROLL_USER_WALLET")
	api.HandleFunc("/wallets/transfer/", walletsHandler.Transfer).Methods("POST").Name("Transfer funds")
	// api.HandleFunc("/operations/", operations.List).Methods("GET").Name("OPERATIONS_LIST")
	r.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)
	r.Handle("/metrics", promhttp.Handler())
	return r
}
