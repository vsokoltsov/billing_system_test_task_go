package app

import (
	"billing_system_test_task/pkg/adapters"
	"billing_system_test_task/pkg/adapters/tx"
	"billing_system_test_task/pkg/entities"
	"billing_system_test_task/pkg/repositories"
	httpHandlers "billing_system_test_task/pkg/transport/http"
	"billing_system_test_task/pkg/usecases"
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
	config entities.ConfigAdapter
	host   string
	port   string
	server *http.Server
	wait   time.Duration
}

func NewApp(config entities.ConfigAdapter) AppAdapter {
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

	usersHandler := httpHandlers.NewUserHandler(userInteractor)
	walletsHandler := httpHandlers.NewWalletsHandler(walletInteractor)
	router := httpHandlers.NewRouter(*usersHandler, *walletsHandler)

	url := strings.Join([]string{host, port}, ":")

	return App{
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
		os.Exit(0)
	}

	log.Println("Shutting down the service...")
	os.Exit(0)
}
