package users

import (
	"billing_system_test_task/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type UsersHandler struct {
	UsersRepo IUserRepo
}

type ErrorMsg struct {
	Message string `json:"message"`
}

// Create godoc
// @Summary Create new user
// @Description Create new user and wallet
// @Tags users
// @Accept  json
// @Produce  json
// @Param user body UserForm true "User attributes"
// @Success 201 {object} UserSerializer "Create user response"
// @Failure 400 {object} formErrorSerializer "User form validation error"
// @Failure default {object} ErrorMsg
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
		json.NewEncoder(w).Encode(formErrorSerializer{Messages: *formError})
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

func (uh *UsersHandler) Enroll(w http.ResponseWriter, r *http.Request) {

}
