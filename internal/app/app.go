package app

import (
	"billing_system_test_task/internal/adapters"
	"billing_system_test_task/internal/adapters/tx"
	"billing_system_test_task/internal/entities"
	"billing_system_test_task/internal/repositories"
	"billing_system_test_task/internal/repositories/reports"
	httpHandlers "billing_system_test_task/internal/transport/http"
	"billing_system_test_task/internal/usecases"
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	_ "github.com/lib/pq"
)

type AppAdapter interface {
	Run()
}

// App represents base application info
type App struct {
	host   string
	port   string
	server *http.Server
	wait   time.Duration
}

func NewApp(config entities.ConfigAdapter) *App {
	var (
		sqlDB        *sql.DB
		sqlDbOpenErr error
		host         = config.GetAppHost()
		port         = config.GetAppPort()
		dbProvider   = config.GetDBProvider()
		dbConnString = config.GetDBConnectionString()
	)

	sqlDB, sqlDbOpenErr = sql.Open(dbProvider, dbConnString)
	if sqlDbOpenErr != nil {
		log.Fatalf("Error sql database open: %s", sqlDbOpenErr)
	}
	if pingErr := sqlDB.Ping(); pingErr != nil {
		log.Fatalf("Error sql database connection: %s", pingErr)
	}
	errFactory := adapters.NewHTTPErrorsFactory()
	txManger := tx.NewTxBeginner(sqlDB)
	walletsRepo := repositories.NewWalletService(sqlDB)
	usersRepo := repositories.NewUsersService(sqlDB)
	operationsRepo := repositories.NewWalletOperationRepo(sqlDB)
	userInteractor := usecases.NewUserInteractor(usersRepo, walletsRepo, operationsRepo, txManger, errFactory)
	walletInteractor := usecases.NewWalletInteractor(walletsRepo, operationsRepo, errFactory, txManger)

	queryParams := reports.NewQueryParamsReader()
	fileStorage := reports.NewFileStorage()
	fileHandler := reports.NewFileHandler(fileStorage)
	pipesManager := reports.NewOperationsProcessesManager()

	operationsInteractor := usecases.NewWalletOperationInteractor(operationsRepo, queryParams, fileHandler, pipesManager, errFactory)

	usersHandler := httpHandlers.NewUserHandler(userInteractor)
	walletsHandler := httpHandlers.NewWalletsHandler(walletInteractor)
	operationsHandler := httpHandlers.NewOperationsHandler(operationsInteractor)
	router := httpHandlers.NewRouter(usersHandler, walletsHandler, operationsHandler)

	url := strings.Join([]string{host, port}, ":")

	return &App{
		host: host,
		port: port,
		wait: time.Second * 5,
		server: &http.Server{
			Handler:      handlers.LoggingHandler(os.Stdout, router),
			Addr:         url,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		},
	}
}

// Run starts application (with gracefull shutdown)
func (a App) Run() {
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
	}

	log.Println("Shutting down the service...")
	os.Exit(0)
}
