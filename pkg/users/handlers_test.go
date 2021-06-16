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
	body           map[string]string
	expectedStatus int
	mockData       func(ctrl *gomock.Controller, ctx context.Context, collection *MockIUserRepo)
	formError      bool
}

var testCases = []userHandlerTestCase{
	userHandlerTestCase{
		name:   "Success user creation",
		method: "POST",
		url:    "/api/users/",
		body: map[string]string{
			"email": "example@mail.com",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, collection *MockIUserRepo) {
			collection.EXPECT().
				Create(ctx, gomock.Any()).
				Return(int64(1), nil)

			collection.EXPECT().
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
		body: map[string]string{
			"email": "example@mail.com",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, collection *MockIUserRepo) {
		},
		expectedStatus: 400,
		formError:      true,
	},
	userHandlerTestCase{
		name:   "Failed user creation (wrong parameters)",
		method: "POST",
		url:    "/api/users/",
		body: map[string]string{
			"email": "test",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, collection *MockIUserRepo) {
		},
		expectedStatus: 400,
	},
	userHandlerTestCase{
		name:   "Failed user creation (failed user repo Create())",
		method: "POST",
		url:    "/api/users/",
		body: map[string]string{
			"email": "example@mail.com",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, collection *MockIUserRepo) {
			collection.EXPECT().
				Create(ctx, gomock.Any()).
				Return(int64(0), fmt.Errorf("User creation error"))
		},
		expectedStatus: 400,
	},
	userHandlerTestCase{
		name:   "Failed user creation (failed user repo GetByID())",
		method: "POST",
		url:    "/api/users/",
		body: map[string]string{
			"email": "example@mail.com",
		},
		mockData: func(ctrl *gomock.Controller, ctx context.Context, collection *MockIUserRepo) {
			collection.EXPECT().
				Create(ctx, gomock.Any()).
				Return(int64(1), nil)

			collection.EXPECT().
				GetByID(ctx, gomock.Any()).
				Return(nil, fmt.Errorf("error of user retrieving"))
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

			mockUsersRepo := NewMockIUserRepo(ctrl)

			r := mux.NewRouter()

			handler := UsersHandler{
				UsersRepo: mockUsersRepo,
			}
			api_router := r.PathPrefix("/api").Subrouter()
			api_router.HandleFunc("/users/", handler.Create).Methods("POST")
			tc.mockData(ctrl, ctx, mockUsersRepo)

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
