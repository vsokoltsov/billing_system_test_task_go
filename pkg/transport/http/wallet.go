package http

import (
	"billing_system_test_task/pkg/repositories"
	"billing_system_test_task/pkg/transport/http/forms"
	"billing_system_test_task/pkg/transport/http/serializers"
	"billing_system_test_task/pkg/usecases"
	"billing_system_test_task/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type WalletsHandler struct {
	WalletRepo    repositories.WalletsManager
	walletUseCase usecases.WalletUseCase
}

func NewWalletsHandler(walletUseCase usecases.WalletUseCase) *WalletsHandler {
	return &WalletsHandler{
		walletUseCase: walletUseCase,
	}
}

// Create godoc
// @Summary Transfer funds
// @Description Transfer funds between two users
// @Tags wallets
// @Accept  json
// @Produce  json
// @Param user body forms.WalletForm true "Transfer parameters"
// @Success 200 {object} serializers.WalletSerializer "Wallet from id"
// @Failure 400 {object} utils.FormErrorSerializer "Wallet transfer validation error"
// @Failure default {object} utils.ErrorMsg
// @Router /api/wallets/transfer/ [post]
func (wh *WalletsHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var (
		walletForm forms.WalletForm
		ctx        = context.Background()
	)
	decoder := json.NewDecoder(r.Body)
	decodeErr := decoder.Decode(&walletForm)
	if decodeErr != nil {
		utils.JsonResponseError(w, http.StatusBadRequest, fmt.Sprintf("Error json form decoding: %s", decodeErr))
		return
	}

	// Validate body parameters
	formError := walletForm.Submit()
	if formError != nil {
		log.Println(fmt.Sprintf("[ERROR] Transfer error - %s", *formError))
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(utils.FormErrorSerializer{Messages: *formError})
		return
	}

	walletFrom, walletTransferErr := wh.walletUseCase.Transfer(ctx, walletForm.WalletFrom, walletForm.WalletTo, walletForm.Amount)
	if walletTransferErr != nil {
		utils.JsonResponseError(w, walletTransferErr.GetStatus(), fmt.Sprintf("Error of funds transfer: %s", walletTransferErr.GetError()))
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(serializers.WalletSerializer{
		WalletFrom: walletFrom,
	})
}
