package http

import (
	"billing_system_test_task/internal/usecases"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
)

// Test defining new routers
func TestAPINewRouter(t *testing.T) {
	ctrl := gomock.NewController(t)
	userUseCase := usecases.NewMockUserUseCase(ctrl)
	walletUseCase := usecases.NewMockWalletUseCase(ctrl)
	operationUseCase := usecases.NewMockWalletOperationUsecase(ctrl)

	userHandler := NewUserHandler(userUseCase)
	walletHandler := NewWalletsHandler(walletUseCase)
	operationHandler := NewOperationsHandler(operationUseCase)

	router := NewRouter(*userHandler, *walletHandler, operationHandler)
	if router == nil {
		t.Error("Expected implementation of http.Handler, got nil")
	}
	_, isMux := router.(*mux.Router)
	if !isMux {
		t.Error("Received instance is not *mux.Router type")
	}
}
