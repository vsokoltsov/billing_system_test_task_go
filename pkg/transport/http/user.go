package http

import (
	"billing_system_test_task/pkg/entities"
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
	"strconv"

	"github.com/gorilla/mux"
)

// UsersHandler stores attributes for handler
type UsersHandler struct {
	UsersRepo   repositories.UsersManager
	WalletsRepo repositories.WalletsManager
	userUseCase usecases.UserUseCase
}

func NewUserHandler(userUseCase usecases.UserUseCase) *UsersHandler {
	return &UsersHandler{
		userUseCase: userUseCase,
	}
}

// Create godoc
// @Summary Create new user
// @Description Create new user and wallet
// @Tags users
// @Accept  json
// @Produce  json
// @Param user body forms.UserForm true "User attributes"
// @Success 201 {object} serializers.UserSerializer "Create user response"
// @Failure 400 {object} utils.FormErrorSerializer "User form validation error"
// @Failure default {object} utils.ErrorMsg
// @Router /api/users/ [post]
func (uh *UsersHandler) Create(w http.ResponseWriter, r *http.Request) {
	var (
		userForm forms.UserForm
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
		log.Println(fmt.Sprintf("[ERROR] Create user: %s", formError))
		w.WriteHeader(http.StatusBadRequest)
		serializer := utils.FormErrorSerializer{Messages: *formError}
		_ = json.NewEncoder(w).Encode(serializer)
		return
	}

	user, createUserErr := uh.userUseCase.Create(ctx, userForm.Email)

	if createUserErr != nil {
		utils.JsonResponseError(w, createUserErr.GetStatus(), createUserErr.GetError().Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(serializers.UserSerializer{
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
// @Param enroll body forms.EnrollForm true "Enrollment attributes"
// @Success 200 {object} serializers.UserSerializer "Retrieving user information with updated balance"
// @Failure 400 {object} utils.FormErrorSerializer "Enroll form validation error"
// @Failure default {object} utils.ErrorMsg
// @Router /api/users/{id}/enroll/ [post]
func (uh *UsersHandler) Enroll(w http.ResponseWriter, r *http.Request) {
	var (
		enrollForm forms.EnrollForm
		ctx        = context.Background()
		user       *entities.User
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
		log.Println(fmt.Sprintf("[ERROR] Enroll user wallet: %s", formError))
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(utils.FormErrorSerializer{Messages: *formError})
		return
	}

	user, walletEnrollErr := uh.userUseCase.Enroll(ctx, userID, enrollForm.Amount)
	if walletEnrollErr != nil {
		utils.JsonResponseError(w, walletEnrollErr.GetStatus(), fmt.Sprintf("Error of wallet enroll: %s", walletEnrollErr.GetError()))
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(serializers.UserSerializer{
		ID:       user.ID,
		Email:    user.Email,
		Balance:  user.Wallet.Balance,
		Currency: user.Wallet.Currency,
	})
}
