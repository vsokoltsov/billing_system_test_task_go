package usecases

import (
	"billing_system_test_task/pkg/adapters"
	"billing_system_test_task/pkg/entities"
	"billing_system_test_task/pkg/repositories"
	"billing_system_test_task/pkg/repositories/reports"
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	reflect "reflect"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	gomock "github.com/golang/mock/gomock"
)

type walletOperationTest struct {
	name                string
	args                []driver.Value
	funcName            string
	mockQuery           func(ctx context.Context, mockOperationsRepo repositories.OperationsManager, mockQueryParams reports.MockQueryReaderManager, mockPipes reports.MockPipelineManager, mockFileMarshaller reports.MockFileMarshallingManager, mockFileHandler reports.MockFileHandlingManager)
	err                 error
	expectedResultMatch func(actual interface{}) bool
}

var walletOperationTests = []walletOperationTest{
	walletOperationTest{
		name:     "Success reports generation",
		funcName: "GenerateReport",
		args:     []driver.Value{},
		mockQuery: func(ctx context.Context, mockOperationsRepo repositories.OperationsManager, mockQueryParams reports.MockQueryReaderManager, mockPipes reports.MockPipelineManager, mockFileMarshaller reports.MockFileMarshallingManager, mockFileHandler reports.MockFileHandlingManager) {
			mu := &sync.Mutex{}
			qp := &reports.QueryParams{
				Format: "json",
			}
			f, _ := os.CreateTemp("", "_example_file")
			fp := &entities.FileParams{
				File: f,
				Path: "_example_file",
				Name: "_example_file",
			}
			fm := reports.NewJSONHandler(
				f,
				mu,
				json.Marshal,
			)
			mockQueryParams.EXPECT().Parse(gomock.Any()).Return(qp, nil)
			mockFileHandler.EXPECT().Create("json").Return(fp, nil)
			mockFileHandler.EXPECT().CreateMarshaller(f, "json", nil).Return(fm, nil)
			mockPipes.EXPECT().Process(ctx, mockOperationsRepo, nil, fm).Return(nil)
			mockFileHandler.EXPECT().GetFileMetadata(f).Return(&entities.Metadata{
				Size:        "100",
				ContentType: "json",
			}, nil)

		},
		expectedResultMatch: func(actual interface{}) bool {
			actualMetadata := actual.(*entities.FileMetadata)
			return actualMetadata != nil
		},
	},
	walletOperationTest{
		name:     "Failed reports generation (query params error)",
		funcName: "GenerateReport",
		args:     []driver.Value{},
		mockQuery: func(ctx context.Context, mockOperationsRepo repositories.OperationsManager, mockQueryParams reports.MockQueryReaderManager, mockPipes reports.MockPipelineManager, mockFileMarshaller reports.MockFileMarshallingManager, mockFileHandler reports.MockFileHandlingManager) {
			mockQueryParams.EXPECT().Parse(gomock.Any()).Return(nil, fmt.Errorf("query params error"))

		},
		err: fmt.Errorf("query params error"),
	},
	walletOperationTest{
		name:     "Failed reports generation (file handler create file error)",
		funcName: "GenerateReport",
		args:     []driver.Value{},
		mockQuery: func(ctx context.Context, mockOperationsRepo repositories.OperationsManager, mockQueryParams reports.MockQueryReaderManager, mockPipes reports.MockPipelineManager, mockFileMarshaller reports.MockFileMarshallingManager, mockFileHandler reports.MockFileHandlingManager) {
			qp := &reports.QueryParams{
				Format: "json",
			}
			mockQueryParams.EXPECT().Parse(gomock.Any()).Return(qp, nil)
			mockFileHandler.EXPECT().Create("json").Return(nil, fmt.Errorf("Create file error"))
		},
		err: fmt.Errorf("Create file error"),
	},
	walletOperationTest{
		name:     "Failed reports generation (create marshaller error)",
		funcName: "GenerateReport",
		args:     []driver.Value{},
		mockQuery: func(ctx context.Context, mockOperationsRepo repositories.OperationsManager, mockQueryParams reports.MockQueryReaderManager, mockPipes reports.MockPipelineManager, mockFileMarshaller reports.MockFileMarshallingManager, mockFileHandler reports.MockFileHandlingManager) {
			qp := &reports.QueryParams{
				Format: "json",
			}
			f, _ := os.CreateTemp("", "_example_file")
			fp := &entities.FileParams{
				File: f,
				Path: "_example_file",
				Name: "_example_file",
			}
			mockQueryParams.EXPECT().Parse(gomock.Any()).Return(qp, nil)
			mockFileHandler.EXPECT().Create("json").Return(fp, nil)
			mockFileHandler.EXPECT().CreateMarshaller(f, "json", nil).Return(nil, fmt.Errorf("create marshaller error"))

		},
		err: fmt.Errorf("create marshaller error"),
	},
	walletOperationTest{
		name:     "Failed reports generation (process error)",
		funcName: "GenerateReport",
		args:     []driver.Value{},
		mockQuery: func(ctx context.Context, mockOperationsRepo repositories.OperationsManager, mockQueryParams reports.MockQueryReaderManager, mockPipes reports.MockPipelineManager, mockFileMarshaller reports.MockFileMarshallingManager, mockFileHandler reports.MockFileHandlingManager) {
			mu := &sync.Mutex{}
			qp := &reports.QueryParams{
				Format: "json",
			}
			f, _ := os.CreateTemp("", "_example_file")
			fp := &entities.FileParams{
				File: f,
				Path: "_example_file",
				Name: "_example_file",
			}
			fm := reports.NewJSONHandler(
				f,
				mu,
				json.Marshal,
			)
			mockQueryParams.EXPECT().Parse(gomock.Any()).Return(qp, nil)
			mockFileHandler.EXPECT().Create("json").Return(fp, nil)
			mockFileHandler.EXPECT().CreateMarshaller(f, "json", nil).Return(fm, nil)
			mockPipes.EXPECT().Process(ctx, mockOperationsRepo, nil, fm).Return(fmt.Errorf("process error"))
		},
		err: fmt.Errorf("process error"),
	},
	walletOperationTest{
		name:     "Failed reports generation (GetFileMetadata error",
		funcName: "GenerateReport",
		args:     []driver.Value{},
		mockQuery: func(ctx context.Context, mockOperationsRepo repositories.OperationsManager, mockQueryParams reports.MockQueryReaderManager, mockPipes reports.MockPipelineManager, mockFileMarshaller reports.MockFileMarshallingManager, mockFileHandler reports.MockFileHandlingManager) {
			mu := &sync.Mutex{}
			qp := &reports.QueryParams{
				Format: "json",
			}
			f, _ := os.CreateTemp("", "_example_file")
			fp := &entities.FileParams{
				File: f,
				Path: "_example_file",
				Name: "_example_file",
			}
			fm := reports.NewJSONHandler(
				f,
				mu,
				json.Marshal,
			)
			mockQueryParams.EXPECT().Parse(gomock.Any()).Return(qp, nil)
			mockFileHandler.EXPECT().Create("json").Return(fp, nil)
			mockFileHandler.EXPECT().CreateMarshaller(f, "json", nil).Return(fm, nil)
			mockPipes.EXPECT().Process(ctx, mockOperationsRepo, nil, fm).Return(nil)
			mockFileHandler.EXPECT().GetFileMetadata(f).Return(nil, fmt.Errorf("metadata error"))

		},
		err: fmt.Errorf("metadata error"),
	},
}

func TestWalletOperationUsecase(t *testing.T) {
	for _, tc := range walletOperationTests {
		ctrl := gomock.NewController(t)
		ctx := context.Background()
		urlVal := make(url.Values)
		realArgs := []reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(urlVal),
		}
		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("cant create mock: %s", err)
		}
		defer db.Close()

		errFactory := adapters.NewHTTPErrorsFactory()
		operationsRepo := repositories.NewMockOperationsManager(ctrl)

		mockQueryParams := reports.NewMockQueryReaderManager(ctrl)
		mockPipes := reports.NewMockPipelineManager(ctrl)
		mockFileMarshaller := reports.NewMockFileMarshallingManager(ctrl)
		mockFileHandler := reports.NewMockFileHandlingManager(ctrl)

		interactor := NewWalletOperationInteractor(operationsRepo, mockQueryParams, mockFileHandler, mockPipes, errFactory).(*WalletOperationInteractor)

		for _, arg := range tc.args {
			realArgs = append(realArgs, reflect.ValueOf(arg))
		}
		tc.mockQuery(ctx, operationsRepo, *mockQueryParams, *mockPipes, *mockFileMarshaller, *mockFileHandler)

		result := reflect.ValueOf(
			interactor,
		).MethodByName(
			tc.funcName,
		).Call(realArgs)
		var (
			reflectErr *adapters.HTTPError
		)
		resultValue := result[0].Interface()
		rerr := result[1].Interface()
		if rerr != nil {
			reflectErr = rerr.(*adapters.HTTPError)
		}

		if reflectErr != nil && tc.err == nil {
			t.Errorf("unexpected err: %s", reflectErr.GetError().Error())
			return
		}

		if tc.err != nil {
			if reflectErr == nil {
				t.Errorf("expected error, got nil: %s", reflectErr.GetError().Error())
				return
			}
			resultErr := rerr.(*adapters.HTTPError)
			if tc.err.Error() != resultErr.GetError().Error() {
				t.Errorf("errors do not match. Expected '%s', got '%s'", tc.err, rerr)
				return
			}
		}

		if tc.err == nil && !tc.expectedResultMatch(resultValue) {
			t.Errorf("result data is not matched. Got %s", resultValue)
		}
	}
}
