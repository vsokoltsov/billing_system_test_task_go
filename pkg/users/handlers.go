package users

import (
	"billing_system_test_task/pkg/utils"
	"billing_system_test_task/pkg/wallets"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type UsersHandler struct {
	UsersRepo   IUserRepo
	WalletsRepo wallets.IWalletRepo
}

// Create godoc
// @Summary Create new user
// @Description Create new user and wallet
// @Tags users
// @Accept  json
// @Produce  json
// @Param user body UserForm true "User attributes"
// @Success 201 {object} UserSerializer "Create user response"
// @Failure 400 {object} utils.FormErrorSerializer "User form validation error"
// @Failure default {object} utils.ErrorMsg
// @Router /api/users/ [post]
func (uh *UsersHandler) Create(w http.ResponseWriter, r *http.Request) {
	var (
		userForm UserForm
		ctx      = context.Background()
	)
	decoder := json.NewDecoder(r.Body)
	decodeErr := decoder.Decode(&userForm)
	if decodeErr != nil {
		utils.JsonResponseError(w, http.StatusBadRequest, fmt.Sprintf("Error json form decoding: %s", decodeErr))
		return
	}

	// Validate body parameters
	formError := userForm.Submit()
	if formError != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.FormErrorSerializer{Messages: *formError})
		return
	}

	userID, createdUserError := uh.UsersRepo.Create(ctx, userForm.Email)
	if createdUserError != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": createdUserError.Error(),
		})
		return
	}

	user, getUserError := uh.UsersRepo.GetByID(ctx, int(userID))
	if getUserError != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": getUserError.Error(),
		})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(UserSerializer{
		ID:       user.ID,
		Email:    user.Email,
		Balance:  user.Wallet.Balance,
		Currency: user.Wallet.Currency,
	})
}

// @Summary Enroll wallet
// @Description Enroll particular users wallet
// @Tags users
// @Accept  json
// @Produce  json
// @Param id path int true "User ID"
// @Param enroll body EnrollForm true "Enrollment attributes"
// @Success 200 {object} UserSerializer "Retrieving user information with updated balance"
// @Failure 400 {object} utils.FormErrorSerializer "Enroll form validation error"
// @Failure default {object} utils.ErrorMsg
// @Router /api/users/{id}/enroll/ [post]
func (uh *UsersHandler) Enroll(w http.ResponseWriter, r *http.Request) {
	var (
		enrollForm EnrollForm
		ctx        = context.Background()
		user       *User
		userGetErr error
	)

	vars := mux.Vars(r)
	userIDVar, userIDExists := vars["id"]
	if !userIDExists {
		utils.JsonResponseError(w, http.StatusInternalServerError, "user's id attribute does not exists")
		return
	}

	userID, errIntConv := strconv.Atoi(userIDVar)
	if errIntConv != nil {
		errorMsg := fmt.Sprintf("Error formatting user id to int: %s", errIntConv)
		utils.JsonResponseError(w, http.StatusBadRequest, errorMsg)
		return
	}

	decoder := json.NewDecoder(r.Body)
	decodeErr := decoder.Decode(&enrollForm)
	if decodeErr != nil {
		utils.JsonResponseError(w, http.StatusBadRequest, fmt.Sprintf("Error json form decoding: %s", decodeErr))
		return
	}

	// Validate body parameters
	formError := enrollForm.Submit()
	if formError != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(utils.FormErrorSerializer{Messages: *formError})
		return
	}

	user, userGetErr = uh.UsersRepo.GetByID(ctx, userID)
	if userGetErr != nil {
		utils.JsonResponseError(w, http.StatusNotFound, userGetErr.Error())
		return
	}

	walletID, walletEnrollErr := uh.WalletsRepo.Enroll(ctx, user.Wallet.ID, enrollForm.Amount)
	if walletEnrollErr != nil {
		utils.JsonResponseError(w, http.StatusBadRequest, fmt.Sprintf("Error of wallet enroll: %s", walletEnrollErr.Error()))
		return
	}

	user, userGetErr = uh.UsersRepo.GetByWalletID(ctx, walletID)
	if userGetErr != nil {
		utils.JsonResponseError(w, http.StatusBadRequest, userGetErr.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(UserSerializer{
		ID:       user.ID,
		Email:    user.Email,
		Balance:  user.Wallet.Balance,
		Currency: user.Wallet.Currency,
	})
}
