package operations

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	gomock "github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

// opHandlerTestCase represents test cases for handler
type opHandlerTestCase struct {
	name           string
	method         string
	url            string
	body           map[string]interface{}
	expectedStatus int
	mockData       func(ctx context.Context, ctrl *gomock.Controller, mock sqlmock.Sqlmock, opRepo *MockOperationsManager, pr *MockQueryReaderManager, fh *MockFileHandlingManager, op *MockPipelineManager)
	formError      bool
	errMsg         string
}

type benchmarkTestCase struct {
	name        string
	url         string
	paramsMap   map[string]string
	queryParams *QueryParams
}

var testCases = []opHandlerTestCase{
	opHandlerTestCase{
		name:           "Success file receiving",
		method:         "GET",
		url:            "/api/operations/?format=csv",
		body:           map[string]interface{}{},
		expectedStatus: 200,
		mockData: func(ctx context.Context, ctrl *gomock.Controller, mock sqlmock.Sqlmock, opRepo *MockOperationsManager, pr *MockQueryReaderManager, fh *MockFileHandlingManager, op *MockPipelineManager) {
			queryParams := make(url.Values)
			queryParams.Set("format", "csv")
			f, _ := os.CreateTemp("", "_example_file")
			params := &QueryParams{
				format:     "csv",
				listParams: nil,
			}
			fileParams := &FileParams{
				f:         f,
				path:      "/a/b/c/" + f.Name(),
				name:      f.Name(),
				csvWriter: csv.NewWriter(f),
			}
			marshaller := &CSVHandler{
				csvWriter: fileParams.csvWriter,
				mu:        &sync.Mutex{},
			}
			query := "select u.id, u.email, w.id, w.user_id, w.balance, w.currency  from users as u"
			rows := sqlmock.NewRows([]string{"id", "email", "wallets.id", "user_id", "balance", "currency"})
			rows = rows.AddRow(1, "test@example.com", 1, 1, decimal.NewFromInt(100), "USD")
			mock.ExpectQuery(query).WithArgs([]driver.Value{1}...).WillReturnRows(rows)

			pr.EXPECT().Parse(queryParams).Return(params, nil)
			fh.EXPECT().Create("csv").Return(fileParams, nil)
			fh.EXPECT().CreateMarshaller(fileParams.f, params.format, fileParams.csvWriter).Return(marshaller, nil)
			op.EXPECT().Process(ctx, opRepo, params.listParams, marshaller).Return(nil)
			fh.EXPECT().GetFileMetadata(fileParams.f).Return(&Metadata{
				size:        "10",
				contentType: "application/json",
			}, nil)
		},
	},
	opHandlerTestCase{
		name:           "Failed file receiving",
		method:         "GET",
		url:            "/api/operations/?format=csv",
		body:           map[string]interface{}{},
		expectedStatus: 400,
		mockData: func(ctx context.Context, ctrl *gomock.Controller, mock sqlmock.Sqlmock, opRepo *MockOperationsManager, pr *MockQueryReaderManager, fh *MockFileHandlingManager, op *MockPipelineManager) {
			queryParams := make(url.Values)
			queryParams.Set("format", "csv")

			pr.EXPECT().Parse(queryParams).Return(nil, fmt.Errorf("query params error"))
		},
		errMsg: "query params error",
	},
	opHandlerTestCase{
		name:           "Failed file receiving (file creation error)",
		method:         "GET",
		url:            "/api/operations/?format=csv",
		body:           map[string]interface{}{},
		expectedStatus: 400,
		mockData: func(ctx context.Context, ctrl *gomock.Controller, mock sqlmock.Sqlmock, opRepo *MockOperationsManager, pr *MockQueryReaderManager, fh *MockFileHandlingManager, op *MockPipelineManager) {
			queryParams := make(url.Values)
			queryParams.Set("format", "csv")
			params := &QueryParams{
				format:     "csv",
				listParams: nil,
			}

			pr.EXPECT().Parse(queryParams).Return(params, nil)
			fh.EXPECT().Create("csv").Return(nil, fmt.Errorf("file create error"))
		},
		errMsg: "file create error",
	},
	opHandlerTestCase{
		name:           "Failed file receiving (file marshaller error)",
		method:         "GET",
		url:            "/api/operations/?format=csv",
		body:           map[string]interface{}{},
		expectedStatus: 400,
		mockData: func(ctx context.Context, ctrl *gomock.Controller, mock sqlmock.Sqlmock, opRepo *MockOperationsManager, pr *MockQueryReaderManager, fh *MockFileHandlingManager, op *MockPipelineManager) {
			queryParams := make(url.Values)
			queryParams.Set("format", "csv")
			f, _ := os.CreateTemp("", "_example_file")
			params := &QueryParams{
				format:     "csv",
				listParams: nil,
			}
			fileParams := &FileParams{
				f:         f,
				path:      "/a/b/c/" + f.Name(),
				name:      f.Name(),
				csvWriter: csv.NewWriter(f),
			}

			pr.EXPECT().Parse(queryParams).Return(params, nil)
			fh.EXPECT().Create("csv").Return(fileParams, nil)
			fh.EXPECT().CreateMarshaller(fileParams.f, "csv", fileParams.csvWriter).Return(nil, fmt.Errorf("file marshaller error"))
		},
		errMsg: "file marshaller error",
	},
	opHandlerTestCase{
		name:           "Failed file receiving (process error)",
		method:         "GET",
		url:            "/api/operations/?format=csv",
		body:           map[string]interface{}{},
		expectedStatus: 400,
		mockData: func(ctx context.Context, ctrl *gomock.Controller, mock sqlmock.Sqlmock, opRepo *MockOperationsManager, pr *MockQueryReaderManager, fh *MockFileHandlingManager, op *MockPipelineManager) {
			queryParams := make(url.Values)
			queryParams.Set("format", "csv")
			f, _ := os.CreateTemp("", "_example_file")
			params := &QueryParams{
				format:     "csv",
				listParams: nil,
			}
			fileParams := &FileParams{
				f:         f,
				path:      "/a/b/c/" + f.Name(),
				name:      f.Name(),
				csvWriter: csv.NewWriter(f),
			}
			marshaller := &CSVHandler{
				csvWriter: fileParams.csvWriter,
				mu:        &sync.Mutex{},
			}
			query := "select u.id, u.email, w.id, w.user_id, w.balance, w.currency  from users as u"
			rows := sqlmock.NewRows([]string{"id", "email", "wallets.id", "user_id", "balance", "currency"})
			rows = rows.AddRow(1, "test@example.com", 1, 1, decimal.NewFromInt(100), "USD")
			mock.ExpectQuery(query).WithArgs([]driver.Value{1}...).WillReturnRows(rows)

			pr.EXPECT().Parse(queryParams).Return(params, nil)
			fh.EXPECT().Create("csv").Return(fileParams, nil)
			fh.EXPECT().CreateMarshaller(fileParams.f, params.format, fileParams.csvWriter).Return(marshaller, nil)
			op.EXPECT().Process(ctx, opRepo, params.listParams, marshaller).Return(fmt.Errorf("process error"))
		},
		errMsg: "process error",
	},
	opHandlerTestCase{
		name:           "Failed file receiving (metadata error)",
		method:         "GET",
		url:            "/api/operations/?format=csv",
		body:           map[string]interface{}{},
		expectedStatus: 400,
		mockData: func(ctx context.Context, ctrl *gomock.Controller, mock sqlmock.Sqlmock, opRepo *MockOperationsManager, pr *MockQueryReaderManager, fh *MockFileHandlingManager, op *MockPipelineManager) {
			queryParams := make(url.Values)
			queryParams.Set("format", "csv")
			f, _ := os.CreateTemp("", "_example_file")
			params := &QueryParams{
				format:     "csv",
				listParams: nil,
			}
			fileParams := &FileParams{
				f:         f,
				path:      "/a/b/c/" + f.Name(),
				name:      f.Name(),
				csvWriter: csv.NewWriter(f),
			}
			marshaller := &CSVHandler{
				csvWriter: fileParams.csvWriter,
				mu:        &sync.Mutex{},
			}
			query := "select u.id, u.email, w.id, w.user_id, w.balance, w.currency  from users as u"
			rows := sqlmock.NewRows([]string{"id", "email", "wallets.id", "user_id", "balance", "currency"})
			rows = rows.AddRow(1, "test@example.com", 1, 1, decimal.NewFromInt(100), "USD")
			mock.ExpectQuery(query).WithArgs([]driver.Value{1}...).WillReturnRows(rows)

			pr.EXPECT().Parse(queryParams).Return(params, nil)
			fh.EXPECT().Create("csv").Return(fileParams, nil)
			fh.EXPECT().CreateMarshaller(fileParams.f, params.format, fileParams.csvWriter).Return(marshaller, nil)
			op.EXPECT().Process(ctx, opRepo, params.listParams, marshaller).Return(nil)
			fh.EXPECT().GetFileMetadata(fileParams.f).Return(nil, fmt.Errorf("metadata error"))
		},
		errMsg: "metadata error",
	},
}

// Test operations package endpoints
func TestOperationsHandler(t *testing.T) {
	for _, tc := range testCases {
		testLabel := strings.Join([]string{"API", tc.method, tc.url, tc.name}, " ")
		t.Run(testLabel, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sqlDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("cant create mock: %s", err)
			}
			defer sqlDB.Close()

			mockOpRepo := NewMockOperationsManager(ctrl)
			mockParamsReader := NewMockQueryReaderManager(ctrl)
			mockFileHandler := NewMockFileHandlingManager(ctrl)
			mockProcessor := NewMockPipelineManager(ctrl)

			r := mux.NewRouter()

			handler := NewOperationsHandler(mockOpRepo, mockParamsReader, mockFileHandler, mockProcessor)
			api_router := r.PathPrefix("/api").Subrouter()
			api_router.HandleFunc("/operations/", handler.List).Methods("GET")
			tc.mockData(ctx, ctrl, mock, mockOpRepo, mockParamsReader, mockFileHandler, mockProcessor)

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

var benchmarks = []benchmarkTestCase{
	benchmarkTestCase{
		name: "Test load for csv format only",
		url:  "/api/operations/?format=csv",
		paramsMap: map[string]string{
			"format": "csv",
		},
		queryParams: &QueryParams{
			format: "csv",
		},
	},
	benchmarkTestCase{
		name: "Test load for reports paging",
		url:  "/api/operations/?page=1&per_page=10",
		paramsMap: map[string]string{
			"page":     "1",
			"per_page": "10",
		},
		queryParams: &QueryParams{
			listParams: &ListParams{
				page:    1,
				perPage: 10,
			},
		},
	},
	benchmarkTestCase{
		name: "Test load for reports date filtering",
		url:  "/api/operations/?date=2020-01-01",
		paramsMap: map[string]string{
			"date": "2020-01-01",
		},
		queryParams: &QueryParams{
			listParams: &ListParams{
				date: "2020-01-01",
			},
		},
	},
	benchmarkTestCase{
		name: "Test load for csv format, date and paging",
		url:  "/api/operations/?format=csv&page=1&per_page=10&date=2020-01-01",
		paramsMap: map[string]string{
			"format":   "csv",
			"page":     "1",
			"per_page": "10",
			"date":     "2020-01-01",
		},
		queryParams: &QueryParams{
			format: "csv",
			listParams: &ListParams{
				page:    1,
				perPage: 10,
				date:    "2020-01-01",
			},
		},
	},
}

// Benchmark /api/operations endpoint with different parameters
func BenchmarkOperationsList(b *testing.B) {
	ctx := context.Background()
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	mockOpRepo := NewMockOperationsManager(ctrl)
	mockParamsReader := NewMockQueryReaderManager(ctrl)
	mockFileHandler := NewMockFileHandlingManager(ctrl)
	mockProcessor := NewMockPipelineManager(ctrl)

	r := mux.NewRouter()

	handler := NewOperationsHandler(mockOpRepo, mockParamsReader, mockFileHandler, mockProcessor)
	api_router := r.PathPrefix("/api").Subrouter()
	api_router.HandleFunc("/operations/", handler.List).Methods("GET")

	f, _ := os.CreateTemp("", "_example_file")
	fileParams := &FileParams{
		f:         f,
		path:      "/a/b/c/" + f.Name(),
		name:      f.Name(),
		csvWriter: csv.NewWriter(f),
	}
	marshaller := &CSVHandler{
		csvWriter: fileParams.csvWriter,
		mu:        &sync.Mutex{},
	}
	query := "select u.id, u.email, w.id, w.user_id, w.balance, w.currency  from users as u"
	rows := sqlmock.NewRows([]string{"id", "email", "wallets.id", "user_id", "balance", "currency"})
	rows = rows.AddRow(1, "test@example.com", 1, 1, decimal.NewFromInt(100), "USD")
	mock.ExpectQuery(query).WithArgs([]driver.Value{1}...).WillReturnRows(rows)

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	body, _ := json.Marshal(map[string]interface{}{})

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			queryParams := make(url.Values)
			for k, v := range bm.paramsMap {
				queryParams.Set(k, v)
			}
			mockFileHandler.EXPECT().Create(bm.queryParams.format).Return(fileParams, nil).AnyTimes()
			mockFileHandler.EXPECT().GetFileMetadata(fileParams.f).Return(&Metadata{
				size:        "10",
				contentType: "application/json",
			}, nil).AnyTimes()
			mockParamsReader.EXPECT().Parse(queryParams).Return(bm.queryParams, nil).AnyTimes()
			mockFileHandler.EXPECT().CreateMarshaller(fileParams.f, bm.queryParams.format, fileParams.csvWriter).Return(marshaller, nil).AnyTimes()
			mockProcessor.EXPECT().Process(ctx, mockOpRepo, bm.queryParams.listParams, marshaller).Return(nil).AnyTimes()

			req, _ := http.NewRequest("GET", testServer.URL+bm.url, bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				r.ServeHTTP(w, req)
			}
		})
	}
}
