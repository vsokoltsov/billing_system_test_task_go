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

type opHandlerTestCase struct {
	name           string
	method         string
	url            string
	body           map[string]interface{}
	expectedStatus int
	mockData       func(ctx context.Context, ctrl *gomock.Controller, mock sqlmock.Sqlmock, opRepo *MockOperationsManager, pr *MockQueryReaderManager, fh *MockFileHandlingManager, op *MockIOperationsProcessesManager)
	formError      bool
	errMsg         string
}

var testCases = []opHandlerTestCase{
	opHandlerTestCase{
		name:           "Success file receiving",
		method:         "GET",
		url:            "/api/operations/?format=csv",
		body:           map[string]interface{}{},
		expectedStatus: 200,
		mockData: func(ctx context.Context, ctrl *gomock.Controller, mock sqlmock.Sqlmock, opRepo *MockOperationsManager, pr *MockQueryReaderManager, fh *MockFileHandlingManager, op *MockIOperationsProcessesManager) {
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
		},
	},
	opHandlerTestCase{
		name:           "Failed file receiving",
		method:         "GET",
		url:            "/api/operations/?format=csv",
		body:           map[string]interface{}{},
		expectedStatus: 400,
		mockData: func(ctx context.Context, ctrl *gomock.Controller, mock sqlmock.Sqlmock, opRepo *MockOperationsManager, pr *MockQueryReaderManager, fh *MockFileHandlingManager, op *MockIOperationsProcessesManager) {
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
		mockData: func(ctx context.Context, ctrl *gomock.Controller, mock sqlmock.Sqlmock, opRepo *MockOperationsManager, pr *MockQueryReaderManager, fh *MockFileHandlingManager, op *MockIOperationsProcessesManager) {
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
		mockData: func(ctx context.Context, ctrl *gomock.Controller, mock sqlmock.Sqlmock, opRepo *MockOperationsManager, pr *MockQueryReaderManager, fh *MockFileHandlingManager, op *MockIOperationsProcessesManager) {
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
		mockData: func(ctx context.Context, ctrl *gomock.Controller, mock sqlmock.Sqlmock, opRepo *MockOperationsManager, pr *MockQueryReaderManager, fh *MockFileHandlingManager, op *MockIOperationsProcessesManager) {
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
}

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
			mockProcessor := NewMockIOperationsProcessesManager(ctrl)

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
