package http

import (
	"billing_system_test_task/internal/adapters"
	"billing_system_test_task/internal/entities"
	"billing_system_test_task/internal/repositories"
	"billing_system_test_task/internal/repositories/reports"
	"billing_system_test_task/internal/usecases"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
)

type operationBenchmark struct {
	name        string
	url         string
	paramsMap   map[string]string
	queryParams *reports.QueryParams
}

type operationWalletTest struct {
	name           string
	method         string
	url            string
	body           map[string]interface{}
	expectedStatus int
	mockData       func(operationUseCase *usecases.MockWalletOperationUsecase)
	formError      bool
	errMsg         string
}

var httpTests = []operationWalletTest{
	operationWalletTest{
		name:   "Success file receiving",
		method: "GET",
		url:    "/api/operations/",
		body: map[string]interface{}{
			"email": "example@mail.com",
		},
		mockData: func(operationUseCase *usecases.MockWalletOperationUsecase) {
			f, _ := os.CreateTemp("", "_example_file")
			operationUseCase.EXPECT().GenerateReport(gomock.Any(), gomock.Any()).Return(&entities.FileMetadata{
				File:        f,
				Path:        "/a/b/c/" + f.Name(),
				Size:        "100",
				ContentType: "json",
			}, nil)
		},
		expectedStatus: 200,
	},
	operationWalletTest{
		name:   "Failed file receiving",
		method: "GET",
		url:    "/api/operations/",
		body: map[string]interface{}{
			"email": "example@mail.com",
		},
		mockData: func(operationUseCase *usecases.MockWalletOperationUsecase) {
			operationUseCase.EXPECT().GenerateReport(gomock.Any(), gomock.Any()).Return(nil, adapters.NewHTTPError(400, fmt.Errorf("generate report error")))
		},
		expectedStatus: 400,
	},
}

// Test operations package endpoints
func TestOperationsHandler(t *testing.T) {
	for _, tc := range httpTests {
		testLabel := strings.Join([]string{"API", tc.method, tc.url, tc.name}, " ")
		t.Run(testLabel, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := mux.NewRouter()

			useCase := usecases.NewMockWalletOperationUsecase(ctrl)
			handler := NewOperationsHandler(useCase)
			api_router := r.PathPrefix("/api").Subrouter()
			api_router.HandleFunc("/operations/", handler.List).Methods("GET")
			tc.mockData(useCase)

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
			if tc.errMsg != "" {
				errors := make(map[string]string)
				respBody, _ := ioutil.ReadAll(resp.Body)
				umErr := json.Unmarshal(respBody, &errors)
				if umErr != nil {
					t.Errorf("Unexpected unmarshalling error: %s", umErr)
				}
				if !strings.Contains(errors["message"], tc.errMsg) {
					t.Errorf("Expect error message '%s'; Got '%s'", tc.errMsg, errors["message"])
				}
			}
		})
	}
}

var operationBenchmarks = []operationBenchmark{
	operationBenchmark{
		name: "Test load for csv format only",
		url:  "/api/operations/?format=csv",
		paramsMap: map[string]string{
			"format": "csv",
		},
		queryParams: &reports.QueryParams{
			Format: "csv",
		},
	},
	operationBenchmark{
		name: "Test load for reports paging",
		url:  "/api/operations/?page=1&per_page=10",
		paramsMap: map[string]string{
			"page":     "1",
			"per_page": "10",
		},
		queryParams: &reports.QueryParams{
			ListParams: &repositories.ListParams{
				Page:    1,
				PerPage: 10,
			},
		},
	},
	operationBenchmark{
		name: "Test load for reports date filtering",
		url:  "/api/operations/?date=2020-01-01",
		paramsMap: map[string]string{
			"date": "2020-01-01",
		},
		queryParams: &reports.QueryParams{
			ListParams: &repositories.ListParams{
				Date: "2020-01-01",
			},
		},
	},
	operationBenchmark{
		name: "Test load for csv format, date and paging",
		url:  "/api/operations/?format=csv&page=1&per_page=10&date=2020-01-01",
		paramsMap: map[string]string{
			"format":   "csv",
			"page":     "1",
			"per_page": "10",
			"date":     "2020-01-01",
		},
		queryParams: &reports.QueryParams{
			Format: "csv",
			ListParams: &repositories.ListParams{
				Page:    1,
				PerPage: 10,
				Date:    "2020-01-01",
			},
		},
	},
}

// Benchmark /api/operations endpoint with different parameters
func BenchmarkOperationsList(b *testing.B) {
	ctx := context.Background()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	r := mux.NewRouter()
	f, _ := os.CreateTemp("", "_example_file")
	useCase := usecases.NewMockWalletOperationUsecase(ctrl)
	useCase.EXPECT().GenerateReport(ctx, gomock.Any()).Return(&entities.FileMetadata{
		File:        f,
		Path:        "/a/b/c/" + f.Name(),
		Size:        "100",
		ContentType: "json",
	}, nil).AnyTimes()
	handler := NewOperationsHandler(useCase)
	api_router := r.PathPrefix("/api").Subrouter()
	api_router.HandleFunc("/operations/", handler.List).Methods("GET")

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	body, _ := json.Marshal(map[string]interface{}{})

	for _, bm := range operationBenchmarks {
		b.Run(bm.name, func(b *testing.B) {
			queryParams := make(url.Values)
			for k, v := range bm.paramsMap {
				queryParams.Set(k, v)
			}

			req, _ := http.NewRequest("GET", testServer.URL+bm.url, bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				r.ServeHTTP(w, req)
			}
		})
	}
}
