package api

import (
	"billing_system_test_task/pkg/operations"
	"billing_system_test_task/pkg/users"
	"billing_system_test_task/pkg/wallets"
	"database/sql"

	"github.com/gorilla/mux"

	_ "billing_system_test_task/docs" // docs is generated by Swag CLI, you have to import it.

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
func SetUpRoutes(pathDelimiter string, sqlDb *sql.DB) *mux.Router {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()

	walletOperationRepo := operations.NewWalletOperationRepo(sqlDb)
	walletsRepo := wallets.NewWalletService(sqlDb, walletOperationRepo)
	usersRepo := users.NewUsersService(sqlDb, walletsRepo)
	users := users.UsersHandler{
		UsersRepo:   usersRepo,
		WalletsRepo: walletsRepo,
	}
	wallets := wallets.WalletsHandler{
		WalletRepo: walletsRepo,
	}
	operations := operations.OperationsHandler{
		OperationsRepo: walletOperationRepo,
	}

	api.HandleFunc("/users/", users.Create).Methods("POST").Name("CREATE_USER")
	api.HandleFunc("/users/{id}/enroll/", users.Enroll).Methods("POST").Name("ENROLL_USER_WALLET")
	api.HandleFunc("/wallets/transfer/", wallets.Transfer).Methods("POST").Name("Transfer funds")
	api.HandleFunc("/operations/", operations.List).Queries("format", "{format}").Methods("GET").Name("OPERATIONS_LIST")
	r.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)
	return r
}
