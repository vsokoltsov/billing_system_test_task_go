package users

import (
	"billing_system_test_task/pkg/wallets"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

type userHandlerTestCase struct {
	name           string
	method         string
	url            string
	body           map[string]interface{}
	expectedStatus int
	mockData       func(ctrl *gomock.Controller, ctx context.Context, userService *MockUsersManager, walletRepo *wallets.MockIWalletRepo)
	formError      bool
}

var testCases = []userHandlerTestCase{
	userHandlerTestCase{
		name:   "Success user creation",
		method: "POST",
		url:    "/api/users/",
		body: map[string]interface{}{
			"email": "example@mail.com",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, userService *MockUsersManager, walletRepo *wallets.MockIWalletRepo) {
			userService.EXPECT().
				Create(ctx, gomock.Any()).
				Return(int64(1), nil)

			userService.EXPECT().
				GetByID(ctx, gomock.Any()).
				Return(
					&User{
						ID:    1,
						Email: "example@mail.com",
						Wallet: &wallets.Wallet{
							Balance:  decimal.NewFromInt(100),
							Currency: "USD",
						},
					},
					nil,
				)
		},
		expectedStatus: 201,
	},
	userHandlerTestCase{
		name:   "Failed user creation (form decode error)",
		method: "POST",
		url:    "/api/users/",
		body: map[string]interface{}{
			"email": "example@mail.com",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, userService *MockUsersManager, walletRepo *wallets.MockIWalletRepo) {
		},
		expectedStatus: 400,
		formError:      true,
	},
	userHandlerTestCase{
		name:   "Failed user creation (wrong parameters)",
		method: "POST",
		url:    "/api/users/",
		body: map[string]interface{}{
			"email": "test",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, userService *MockUsersManager, walletRepo *wallets.MockIWalletRepo) {
		},
		expectedStatus: 400,
	},
	userHandlerTestCase{
		name:   "Failed user creation (failed user repo Create())",
		method: "POST",
		url:    "/api/users/",
		body: map[string]interface{}{
			"email": "example@mail.com",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, userService *MockUsersManager, walletRepo *wallets.MockIWalletRepo) {
			userService.EXPECT().
				Create(ctx, gomock.Any()).
				Return(int64(0), fmt.Errorf("User creation error"))
		},
		expectedStatus: 400,
	},
	userHandlerTestCase{
		name:   "Failed user creation (failed user repo GetByID())",
		method: "POST",
		url:    "/api/users/",
		body: map[string]interface{}{
			"email": "example@mail.com",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, userService *MockUsersManager, walletRepo *wallets.MockIWalletRepo) {
			userService.EXPECT().
				Create(ctx, gomock.Any()).
				Return(int64(1), nil)

			userService.EXPECT().
				GetByID(ctx, gomock.Any()).
				Return(nil, fmt.Errorf("error of user retrieving"))
		},
		expectedStatus: 400,
	},
	userHandlerTestCase{
		name:   "Success wallet enroll",
		method: "POST",
		url:    "/api/users/1/enroll/",
		body: map[string]interface{}{
			"amount": "100",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, userService *MockUsersManager, walletRepo *wallets.MockIWalletRepo) {
			amount := decimal.NewFromInt(100)
			user := User{
				ID:    1,
				Email: "example@mail.com",
				Wallet: &wallets.Wallet{
					ID:       1,
					UserID:   1,
					Balance:  decimal.NewFromInt(0),
					Currency: "USD",
				},
			}
			userService.EXPECT().GetByID(ctx, user.ID).Return(&user, nil)

			walletRepo.EXPECT().Enroll(ctx, user.Wallet.ID, amount).Return(user.Wallet.ID, nil)

			user.Wallet.Balance = user.Wallet.Balance.Add(amount)

			userService.EXPECT().GetByWalletID(ctx, user.Wallet.ID).Return(&user, nil)
		},
		expectedStatus: 200,
	},
	userHandlerTestCase{
		name:   "Failed wallet enroll (vars parameter does not exists)",
		method: "POST",
		url:    "/api/users/1/enroll/",
		body: map[string]interface{}{
			"amount": "100",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, userService *MockUsersManager, walletRepo *wallets.MockIWalletRepo) {
		},
		expectedStatus: 500,
	},
	userHandlerTestCase{
		name:   "Failed wallet enroll (error of user id to int conversion)",
		method: "POST",
		url:    "/api/users/test/enroll/",
		body: map[string]interface{}{
			"amount": "100",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, userService *MockUsersManager, walletRepo *wallets.MockIWalletRepo) {
		},
		expectedStatus: 400,
	},
	userHandlerTestCase{
		name:   "Failed wallet enroll (form decoding error)",
		method: "POST",
		url:    "/api/users/1/enroll/",
		body: map[string]interface{}{
			"amount": "100",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, userService *MockUsersManager, walletRepo *wallets.MockIWalletRepo) {
		},
		expectedStatus: 400,
		formError:      true,
	},
	userHandlerTestCase{
		name:   "Failed wallet enroll (form validation error)",
		method: "POST",
		url:    "/api/users/1/enroll/",
		body: map[string]interface{}{
			"amount": 0,
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, userService *MockUsersManager, walletRepo *wallets.MockIWalletRepo) {
		},
		expectedStatus: 400,
	},
	userHandlerTestCase{
		name:   "Failed wallet enroll (user not found)",
		method: "POST",
		url:    "/api/users/1/enroll/",
		body: map[string]interface{}{
			"amount": "100",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, userService *MockUsersManager, walletRepo *wallets.MockIWalletRepo) {
			userService.EXPECT().GetByID(ctx, 1).Return(nil, fmt.Errorf("user not found"))
		},
		expectedStatus: 404,
	},
	userHandlerTestCase{
		name:   "Failed wallet enroll (wallet repo Enroll() failed)",
		method: "POST",
		url:    "/api/users/1/enroll/",
		body: map[string]interface{}{
			"amount": "100",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, userService *MockUsersManager, walletRepo *wallets.MockIWalletRepo) {
			amount := decimal.NewFromInt(100)
			user := User{
				ID:    1,
				Email: "example@mail.com",
				Wallet: &wallets.Wallet{
					ID:       1,
					UserID:   1,
					Balance:  decimal.NewFromInt(0),
					Currency: "USD",
				},
			}
			userService.EXPECT().GetByID(ctx, user.ID).Return(&user, nil)

			walletRepo.EXPECT().Enroll(ctx, user.Wallet.ID, amount).Return(0, fmt.Errorf("enroll has failed"))
		},
		expectedStatus: 400,
	},
	userHandlerTestCase{
		name:   "Failed wallet enroll (failed user repo GetByWalletID())",
		method: "POST",
		url:    "/api/users/1/enroll/",
		body: map[string]interface{}{
			"amount": "100",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, userService *MockUsersManager, walletRepo *wallets.MockIWalletRepo) {
			amount := decimal.NewFromInt(100)
			user := User{
				ID:    1,
				Email: "example@mail.com",
				Wallet: &wallets.Wallet{
					ID:       1,
					UserID:   1,
					Balance:  decimal.NewFromInt(0),
					Currency: "USD",
				},
			}
			userService.EXPECT().GetByID(ctx, user.ID).Return(&user, nil)

			walletRepo.EXPECT().Enroll(ctx, user.Wallet.ID, amount).Return(user.Wallet.ID, nil)

			user.Wallet.Balance = user.Wallet.Balance.Add(amount)

			userService.EXPECT().GetByWalletID(ctx, user.Wallet.ID).Return(nil, fmt.Errorf("error of user retrieving"))
		},
		expectedStatus: 400,
	},
}

func TestUsersHandlers(t *testing.T) {
	for _, tc := range testCases {
		testLabel := strings.Join([]string{"API", tc.method, tc.url, tc.name}, " ")
		t.Run(testLabel, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sqlDB, _, err := sqlmock.New()
			if err != nil {
				t.Fatalf("cant create mock: %s", err)
			}
			defer sqlDB.Close()

			mockUsersRepo := NewMockUsersManager(ctrl)
			mockWalletsRepo := wallets.NewMockIWalletRepo(ctrl)

			enrollRoute := "/users/{id}/enroll/"
			if tc.name == "Failed wallet enroll (vars parameter does not exists)" {
				enrollRoute = "/users/{test}/enroll/"
			}

			r := mux.NewRouter()

			handler := UsersHandler{
				UsersRepo:   mockUsersRepo,
				WalletsRepo: mockWalletsRepo,
			}
			api_router := r.PathPrefix("/api").Subrouter()
			api_router.HandleFunc("/users/", handler.Create).Methods("POST")
			api_router.HandleFunc(enrollRoute, handler.Enroll).Methods("POST")
			tc.mockData(ctrl, ctx, mockUsersRepo, mockWalletsRepo)

			testServer := httptest.NewServer(r)
			defer testServer.Close()

			var body []byte
			if tc.formError {
				body = []byte(`{"test": "data"`)
			} else {
				body, _ = json.Marshal(tc.body)
			}

			req, _ := http.NewRequest(tc.method, testServer.URL+tc.url, bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("[%s] Expected response code %d. Got %d", testLabel, tc.expectedStatus, resp.StatusCode)
			}
		})
	}
}
