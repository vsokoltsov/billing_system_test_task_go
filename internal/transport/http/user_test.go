package http

import (
	"billing_system_test_task/internal/adapters"
	"billing_system_test_task/internal/entities"
	"billing_system_test_task/internal/transport/http/serializers"
	"billing_system_test_task/internal/usecases"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

// userHandlerTestCase store test cases information
type userHandlerTestCase struct {
	name           string
	method         string
	url            string
	body           map[string]interface{}
	expectedStatus int
	mockData       func(userUsecase *usecases.MockUserUseCase)
	formError      bool
	matchResults   func(actual []byte) bool
}

var (
	createUser = userHandlerTestCase{
		name:   "Success user creation",
		method: "POST",
		url:    "/api/users/",
		body: map[string]interface{}{
			"email": "example@mail.com",
		},
		mockData: func(userUsecase *usecases.MockUserUseCase) {
			user := &entities.User{
				ID:    1,
				Email: "example@mail.com",
				Wallet: &entities.Wallet{
					Balance:  decimal.NewFromInt(100),
					Currency: "USD",
				},
			}
			userUsecase.EXPECT().Create(gomock.Any(), "example@mail.com").Return(
				user,
				nil,
			).AnyTimes()
		},
		expectedStatus: 201,
		matchResults: func(actual []byte) bool {
			var serializer serializers.UserSerializer
			_ = json.Unmarshal(actual, &serializer)
			return serializer.ID == 1 && serializer.Email == "example@mail.com" && serializer.Balance.IntPart() == int64(100) && serializer.Currency == "USD"
		},
	}
	enroll = userHandlerTestCase{
		name:   "Success wallet enroll",
		method: "POST",
		url:    "/api/users/1/enroll/",
		body: map[string]interface{}{
			"amount": "100",
		},
		mockData: func(userUsecase *usecases.MockUserUseCase) {
			user := entities.User{
				ID:    1,
				Email: "example@mail.com",
				Wallet: &entities.Wallet{
					ID:       1,
					UserID:   1,
					Balance:  decimal.NewFromInt(100),
					Currency: "USD",
				},
			}
			userUsecase.EXPECT().Enroll(gomock.Any(), 1, decimal.NewFromInt(100)).Return(
				&user,
				nil,
			).AnyTimes()
		},
		expectedStatus: 200,
		matchResults: func(actual []byte) bool {
			var serializer serializers.UserSerializer
			_ = json.Unmarshal(actual, &serializer)
			return serializer.ID == 1 && serializer.Email == "example@mail.com" && serializer.Balance.IntPart() == int64(100) && serializer.Currency == "USD"
		},
	}
)

var testCases = []userHandlerTestCase{
	createUser,
	userHandlerTestCase{
		name:   "Failed user creation (form decode error)",
		method: "POST",
		url:    "/api/users/",
		body: map[string]interface{}{
			"email": "example@mail.com",
		},
		mockData: func(userUsecase *usecases.MockUserUseCase) {
		},
		expectedStatus: 400,
		formError:      true,
		matchResults: func(actual []byte) bool {
			var errors ErrorMsg
			_ = json.Unmarshal(actual, &errors)
			return strings.Contains(errors.Message, "unexpected EOF")
		},
	},
	userHandlerTestCase{
		name:   "Failed user creation (wrong parameters)",
		method: "POST",
		url:    "/api/users/",
		body: map[string]interface{}{
			"email": "test",
		},
		mockData: func(userUsecase *usecases.MockUserUseCase) {
		},
		expectedStatus: 400,
		matchResults: func(actual []byte) bool {
			var errors FormErrorSerializer
			_ = json.Unmarshal(actual, &errors)
			return errors.Messages["email"][0] == "Invalid email format"
		},
	},
	userHandlerTestCase{
		name:   "Failed user creation (failed user repo Create())",
		method: "POST",
		url:    "/api/users/",
		body: map[string]interface{}{
			"email": "example@mail.com",
		},
		mockData: func(userUsecase *usecases.MockUserUseCase) {
			userUsecase.EXPECT().Create(gomock.Any(), "example@mail.com").Return(
				nil,
				adapters.NewHTTPError(400, fmt.Errorf("User creation error")),
			).AnyTimes()
		},
		expectedStatus: 400,
		matchResults: func(actual []byte) bool {
			var errors ErrorMsg
			_ = json.Unmarshal(actual, &errors)
			return strings.Contains(errors.Message, "User creation error")
		},
	},
	userHandlerTestCase{
		name:   "Failed user creation (failed user repo GetByID())",
		method: "POST",
		url:    "/api/users/",
		body: map[string]interface{}{
			"email": "example@mail.com",
		},
		mockData: func(userUsecase *usecases.MockUserUseCase) {
			userUsecase.EXPECT().Create(gomock.Any(), "example@mail.com").Return(
				nil,
				adapters.NewHTTPError(404, fmt.Errorf("error of user retrieving")),
			).AnyTimes()
		},
		expectedStatus: 404,
		matchResults: func(actual []byte) bool {
			var errors ErrorMsg
			_ = json.Unmarshal(actual, &errors)
			return strings.Contains(errors.Message, "error of user retrieving")
		},
	},
	enroll,
	userHandlerTestCase{
		name:   "Failed wallet enroll (vars parameter does not exists)",
		method: "POST",
		url:    "/api/users/1/enroll/",
		body: map[string]interface{}{
			"amount": "100",
		},
		mockData: func(userUsecase *usecases.MockUserUseCase) {
		},
		expectedStatus: 500,
		matchResults: func(actual []byte) bool {
			var errors ErrorMsg
			_ = json.Unmarshal(actual, &errors)
			return strings.Contains(errors.Message, "user's id attribute does not exists")
		},
	},
	userHandlerTestCase{
		name:   "Failed wallet enroll (error of user id to int conversion)",
		method: "POST",
		url:    "/api/users/test/enroll/",
		body: map[string]interface{}{
			"amount": "100",
		},
		mockData: func(userUsecase *usecases.MockUserUseCase) {
		},
		expectedStatus: 400,
		matchResults: func(actual []byte) bool {
			var errors ErrorMsg
			_ = json.Unmarshal(actual, &errors)
			return strings.Contains(errors.Message, "Error formatting user id to int")
		},
	},
	userHandlerTestCase{
		name:   "Failed wallet enroll (form decoding error)",
		method: "POST",
		url:    "/api/users/1/enroll/",
		body: map[string]interface{}{
			"amount": "100",
		},
		mockData: func(userUsecase *usecases.MockUserUseCase) {
		},
		expectedStatus: 400,
		formError:      true,
		matchResults: func(actual []byte) bool {
			var errors ErrorMsg
			_ = json.Unmarshal(actual, &errors)
			return strings.Contains(errors.Message, "Error json form decoding")
		},
	},
	userHandlerTestCase{
		name:   "Failed wallet enroll (form validation error)",
		method: "POST",
		url:    "/api/users/1/enroll/",
		body: map[string]interface{}{
			"amount": 0,
		},
		mockData: func(userUsecase *usecases.MockUserUseCase) {
		},
		expectedStatus: 400,
		matchResults: func(actual []byte) bool {
			var errors FormErrorSerializer
			_ = json.Unmarshal(actual, &errors)
			return errors.Messages["amount"][0] == "less than a zero"
		},
	},
	userHandlerTestCase{
		name:   "Failed wallet enroll (user not found)",
		method: "POST",
		url:    "/api/users/1/enroll/",
		body: map[string]interface{}{
			"amount": "100",
		},
		mockData: func(userUsecase *usecases.MockUserUseCase) {
			userUsecase.EXPECT().Enroll(gomock.Any(), 1, decimal.NewFromInt(100)).Return(
				nil,
				adapters.NewHTTPError(404, fmt.Errorf("user not found")),
			).AnyTimes()
		},
		expectedStatus: 404,
		matchResults: func(actual []byte) bool {
			var errors ErrorMsg
			_ = json.Unmarshal(actual, &errors)
			return strings.Contains(errors.Message, "user not found")
		},
	},
	userHandlerTestCase{
		name:   "Failed wallet enroll (wallet repo Enroll() failed)",
		method: "POST",
		url:    "/api/users/1/enroll/",
		body: map[string]interface{}{
			"amount": "100",
		},
		mockData: func(userUsecase *usecases.MockUserUseCase) {
			userUsecase.EXPECT().Enroll(gomock.Any(), 1, decimal.NewFromInt(100)).Return(
				nil,
				adapters.NewHTTPError(400, fmt.Errorf("enroll has failed")),
			).AnyTimes()
		},
		expectedStatus: 400,
		matchResults: func(actual []byte) bool {
			var errors ErrorMsg
			_ = json.Unmarshal(actual, &errors)
			return strings.Contains(errors.Message, "enroll has failed")
		},
	},
	userHandlerTestCase{
		name:   "Failed wallet enroll (failed user repo GetByWalletID())",
		method: "POST",
		url:    "/api/users/1/enroll/",
		body: map[string]interface{}{
			"amount": "100",
		},
		mockData: func(userUsecase *usecases.MockUserUseCase) {
			userUsecase.EXPECT().Enroll(gomock.Any(), 1, decimal.NewFromInt(100)).Return(
				nil,
				adapters.NewHTTPError(404, fmt.Errorf("error of user retrieving")),
			).AnyTimes()
		},
		expectedStatus: 404,
		matchResults: func(actual []byte) bool {
			var errors ErrorMsg
			_ = json.Unmarshal(actual, &errors)
			return strings.Contains(errors.Message, "error of user retrieving")
		},
	},
}

// Tests users' handlers
func TestUsersHandlers(t *testing.T) {
	for _, tc := range testCases {
		testLabel := strings.Join([]string{"API", tc.method, tc.url, tc.name}, " ")
		t.Run(testLabel, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sqlDB, _, err := sqlmock.New()
			if err != nil {
				t.Fatalf("cant create mock: %s", err)
			}
			defer sqlDB.Close()

			enrollRoute := "/users/{id}/enroll/"
			if tc.name == "Failed wallet enroll (vars parameter does not exists)" {
				enrollRoute = "/users/{test}/enroll/"
			}

			r := mux.NewRouter()
			interactor := usecases.NewMockUserUseCase(ctrl)
			handler := NewUserHandler(interactor)
			api_router := r.PathPrefix("/api").Subrouter()
			api_router.HandleFunc("/users/", handler.Create).Methods("POST")
			api_router.HandleFunc(enrollRoute, handler.Enroll).Methods("POST")
			tc.mockData(interactor)

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
			respBody, _ := ioutil.ReadAll(resp.Body)

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("[%s] Expected response code %d. Got %d", testLabel, tc.expectedStatus, resp.StatusCode)
			}

			if !tc.matchResults(respBody) {
				t.Errorf("[%s] Unmatched results. Got %s", testLabel, string(respBody))
			}
		})
	}
}

var benchmarks = []userHandlerTestCase{
	createUser,
	enroll,
}

// Benchmark users' handlers
func BenchmarkUsers(b *testing.B) {
	for _, tc := range benchmarks {
		testLabel := strings.Join([]string{"API", tc.method, tc.url, tc.name}, " ")
		b.Run(testLabel, func(b *testing.B) {
			ctrl := gomock.NewController(b)
			defer ctrl.Finish()

			sqlDB, _, err := sqlmock.New()
			if err != nil {
				b.Fatalf("cant create mock: %s", err)
			}
			defer sqlDB.Close()
			interactor := usecases.NewMockUserUseCase(ctrl)

			enrollRoute := "/users/{id}/enroll/"
			if tc.name == "Failed wallet enroll (vars parameter does not exists)" {
				enrollRoute = "/users/{test}/enroll/"
			}

			r := mux.NewRouter()

			handler := NewUserHandler(interactor)
			api_router := r.PathPrefix("/api").Subrouter()
			api_router.HandleFunc("/users/", handler.Create).Methods("POST")
			api_router.HandleFunc(enrollRoute, handler.Enroll).Methods("POST")
			tc.mockData(interactor)

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
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				r.ServeHTTP(w, req)
			}
		})
	}
}
