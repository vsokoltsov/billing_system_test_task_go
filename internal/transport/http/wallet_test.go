package http

import (
	"billing_system_test_task/internal/adapters"
	"billing_system_test_task/internal/transport/http/serializers"
	"billing_system_test_task/internal/usecases"
	"billing_system_test_task/internal/utils"
	"bytes"
	"context"
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

// walletHandlerTestCase stores data for wallet handler tests
type walletHandlerTestCase struct {
	name           string
	method         string
	url            string
	body           map[string]interface{}
	expectedStatus int
	mockData       func(ctx context.Context, walletUseCase *usecases.MockWalletUseCase)
	matchResults   func(actual []byte) bool
	formError      bool
}

var (
	transfer = walletHandlerTestCase{
		name:   "Success funds transfering",
		method: "POST",
		url:    "/api/wallets/transfer/",
		body: map[string]interface{}{
			"wallet_from": 1,
			"wallet_to":   2,
			"amount":      decimal.NewFromInt(25),
		},
		mockData: func(ctx context.Context, walletUseCase *usecases.MockWalletUseCase) {
			walletUseCase.EXPECT().Transfer(ctx, 1, 2, decimal.NewFromInt(25)).Return(1, nil)
		},
		expectedStatus: 200,
		matchResults: func(actual []byte) bool {
			var ws serializers.WalletSerializer
			_ = json.Unmarshal(actual, &ws)
			return ws.WalletFrom == 1
		},
	}
)

var walletTestCases = []walletHandlerTestCase{
	transfer,
	walletHandlerTestCase{
		name:   "Failed funds transfering (form decoding error)",
		method: "POST",
		url:    "/api/wallets/transfer/",
		body: map[string]interface{}{
			"wallet_from": 1,
			"wallet_to":   2,
			"amount":      decimal.NewFromInt(25),
		},
		mockData: func(ctx context.Context, walletUseCase *usecases.MockWalletUseCase) {
		},
		expectedStatus: 400,
		formError:      true,
		matchResults: func(actual []byte) bool {
			var errors utils.ErrorMsg
			_ = json.Unmarshal(actual, &errors)
			return strings.Contains(errors.Message, "unexpected EOF")
		},
	},
	walletHandlerTestCase{
		name:   "Failed funds transfering (form validation error)",
		method: "POST",
		url:    "/api/wallets/transfer/",
		body: map[string]interface{}{
			"wallet_from": 1,
		},
		mockData: func(ctx context.Context, walletUseCase *usecases.MockWalletUseCase) {
		},
		expectedStatus: 400,
		matchResults: func(actual []byte) bool {
			var errors utils.FormErrorSerializer
			_ = json.Unmarshal(actual, &errors)
			return errors.Messages["amount"][0] == "less than a zero"
		},
	},
	walletHandlerTestCase{
		name:   "Failed funds transfering (funds transfer error)",
		method: "POST",
		url:    "/api/wallets/transfer/",
		body: map[string]interface{}{
			"wallet_from": 1,
			"wallet_to":   2,
			"amount":      decimal.NewFromInt(25),
		},
		mockData: func(ctx context.Context, walletUseCase *usecases.MockWalletUseCase) {
			walletUseCase.EXPECT().Transfer(ctx, 1, 2, decimal.NewFromInt(25)).Return(0, adapters.NewHTTPError(400, fmt.Errorf("Error of funds transfering")))
		},
		expectedStatus: 400,
		matchResults: func(actual []byte) bool {
			var errors utils.ErrorMsg
			_ = json.Unmarshal(actual, &errors)
			return strings.Contains(errors.Message, "Error of funds transfering")
		},
	},
	walletHandlerTestCase{
		name:   "Failed funds transfering (walletFrom and walletTo are equal)",
		method: "POST",
		url:    "/api/wallets/transfer/",
		body: map[string]interface{}{
			"wallet_from": 1,
			"wallet_to":   1,
			"amount":      decimal.NewFromInt(25),
		},
		mockData: func(ctx context.Context, walletUseCase *usecases.MockWalletUseCase) {
			// walletUseCase.EXPECT().Transfer(ctx, 1, 2, decimal.NewFromInt(25)).Return(0, adapters.NewHTTPError(400, fmt.Errorf("Error of funds transfering")))
		},
		expectedStatus: 400,
		matchResults: func(actual []byte) bool {
			var errors utils.FormErrorSerializer
			_ = json.Unmarshal(actual, &errors)
			return errors.Messages["wallet_from"][0] == "source wallet is equal to destination wallet"
		},
	},
}

// // Test wallets handlers
func TestWalletHandlers(t *testing.T) {
	for _, tc := range walletTestCases {
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

			mockWalletUseCase := usecases.NewMockWalletUseCase(ctrl)

			r := mux.NewRouter()

			handler := NewWalletsHandler(mockWalletUseCase)
			api_router := r.PathPrefix("/api").Subrouter()
			api_router.HandleFunc("/wallets/transfer/", handler.Transfer).Methods("POST")
			tc.mockData(ctx, mockWalletUseCase)

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

var walletBenchmarks = []walletHandlerTestCase{
	transfer,
}

// Benchmarks wallets handlers
func BenchmarkWalletsHandler(b *testing.B) {
	for _, tc := range walletBenchmarks {
		testLabel := strings.Join([]string{"API", tc.method, tc.url, tc.name}, " ")
		b.Run(testLabel, func(b *testing.B) {
			ctx := context.Background()
			ctrl := gomock.NewController(b)
			defer ctrl.Finish()

			sqlDB, _, err := sqlmock.New()
			if err != nil {
				b.Fatalf("cant create mock: %s", err)
			}
			defer sqlDB.Close()

			mockWalletUseCase := usecases.NewMockWalletUseCase(ctrl)

			r := mux.NewRouter()

			handler := NewWalletsHandler(mockWalletUseCase)
			api_router := r.PathPrefix("/api").Subrouter()
			api_router.HandleFunc("/wallets/transfer/", handler.Transfer).Methods("POST")
			tc.mockData(ctx, mockWalletUseCase)

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
