package repositories

import (
	"billing_system_test_task/pkg/adapters/tx"
	"billing_system_test_task/pkg/entities"
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
)

// operationRepoTestCase represents test cases for repository
type operationRepoTestCase struct {
	name                string
	funcName            string
	mockQuery           func(mock sqlmock.Sqlmock)
	args                []driver.Value
	err                 error
	expectedResultMatch func(actual interface{}) bool
}

var operationRepoTestCases = []operationRepoTestCase{
	operationRepoTestCase{
		name:     "Success operation creation",
		funcName: "Create",
		args:     []driver.Value{Create, 1, 2, decimal.NewFromInt(0)},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id"})
			rows = rows.AddRow(1)

			// Exec insert wallets
			mock.
				ExpectQuery("insert into wallet_operations").
				WithArgs([]driver.Value{Create, 1, 2, decimal.NewFromInt(0)}...).
				WillReturnRows(rows)

		},
		expectedResultMatch: func(actual interface{}) bool {
			return actual.(int) == 1
		},
	},
	operationRepoTestCase{
		name:     "Success operation creation (walletFrom is 0)",
		funcName: "Create",
		args:     []driver.Value{Create, 0, 2, decimal.NewFromInt(0)},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id"})
			rows = rows.AddRow(1)

			// Exec insert wallets
			mock.
				ExpectQuery("insert into wallet_operations").
				WithArgs([]driver.Value{Create, nil, 2, decimal.NewFromInt(0)}...).
				WillReturnRows(rows)

		},
		expectedResultMatch: func(actual interface{}) bool {
			return actual.(int) == 1
		},
	},
	operationRepoTestCase{
		name:     "Failed wallet creation (insert error)",
		funcName: "Create",
		args:     []driver.Value{Create, 1, 2, decimal.NewFromInt(0)},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Exec insert operation
			mock.
				ExpectQuery("insert into wallet_operations").
				WithArgs([]driver.Value{Create, 1, 2, decimal.NewFromInt(0)}...).
				WillReturnError(fmt.Errorf("Insert error"))
		},
		err: fmt.Errorf("Insert error"),
	},
	operationRepoTestCase{
		name:     "Failed wallet creation (Scan error)",
		funcName: "Create",
		args:     []driver.Value{Create, 1, 2, decimal.NewFromInt(0)},
		mockQuery: func(mock sqlmock.Sqlmock) {
			rows := sqlmock.NewRows([]string{"id"})
			rows = rows.AddRow(nil).RowError(1, fmt.Errorf("Scan error"))

			// Exec insert operation
			mock.
				ExpectQuery("insert into wallet_operations").
				WithArgs([]driver.Value{Create, 1, 2, decimal.NewFromInt(0)}...).
				WillReturnRows(rows)

		},
		err: fmt.Errorf("Scan error"),
	},
	operationRepoTestCase{
		name:     "Success receiving list of items",
		funcName: "List",
		args:     []driver.Value{nil},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})
			rows = rows.AddRow(1, Create, nil, 1, decimal.NewFromInt(0), time.Now())

			// Exec insert wallets
			mock.
				ExpectQuery("select id, operation, wallet_from, wallet_to, amount, created_at from wallet_operations").
				WillReturnRows(rows)

		},
		expectedResultMatch: func(actual interface{}) bool {
			rows := actual.(chan *entities.WalletOperation)
			operation := <-rows
			return operation.ID == 1
		},
	},
	operationRepoTestCase{
		name:     "Success receiving list of items with paging",
		funcName: "List",
		args:     []driver.Value{&ListParams{Page: 1, PerPage: 10}},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})
			rows = rows.AddRow(1, Create, nil, 1, decimal.NewFromInt(0), time.Now())

			// Exec insert wallets
			mock.
				ExpectQuery("select id, operation, wallet_from, wallet_to, amount, created_at from wallet_operations").
				WithArgs([]driver.Value{0, 10}...).
				WillReturnRows(rows)

		},
		expectedResultMatch: func(actual interface{}) bool {
			rows := actual.(chan *entities.WalletOperation)
			operation := <-rows
			return operation.ID == 1
		},
	},
	operationRepoTestCase{
		name:     "Success receiving list of items with paging more than 1",
		funcName: "List",
		args:     []driver.Value{&ListParams{Page: 3, PerPage: 10}},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})
			rows = rows.AddRow(1, Create, nil, 1, decimal.NewFromInt(0), time.Now())

			// Exec insert wallets
			mock.
				ExpectQuery("select id, operation, wallet_from, wallet_to, amount, created_at from wallet_operations").
				WithArgs([]driver.Value{20, 10}...).
				WillReturnRows(rows)

		},
		expectedResultMatch: func(actual interface{}) bool {
			rows := actual.(chan *entities.WalletOperation)
			operation := <-rows
			return operation.ID == 1
		},
	},
	operationRepoTestCase{
		name:     "Success receiving list of items date filtering",
		funcName: "List",
		args:     []driver.Value{&ListParams{Date: "2020-01-01"}},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})
			rows = rows.AddRow(1, Create, nil, 1, decimal.NewFromInt(0), time.Now())

			// Exec insert wallets
			mock.
				ExpectQuery("where created_at = to_date").
				WithArgs([]driver.Value{"2020-01-01"}...).
				WillReturnRows(rows)

		},
		expectedResultMatch: func(actual interface{}) bool {
			rows := actual.(chan *entities.WalletOperation)
			operation := <-rows
			return operation.ID == 1
		},
	},
	operationRepoTestCase{
		name:     "Success receiving list of items with all parameters",
		funcName: "List",
		args:     []driver.Value{&ListParams{Page: 1, PerPage: 10, Date: "2020-01-01"}},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})
			rows = rows.AddRow(1, Create, nil, 1, decimal.NewFromInt(0), time.Now())

			// Exec insert wallets
			mock.
				ExpectQuery("where created_at = to_date").
				WithArgs([]driver.Value{"2020-01-01", 0, 10}...).
				WillReturnRows(rows)

		},
		expectedResultMatch: func(actual interface{}) bool {
			rows := actual.(chan *entities.WalletOperation)
			operation := <-rows
			return operation.ID == 1
		},
	},
	operationRepoTestCase{
		name:     "Success receiving of empty list",
		funcName: "List",
		args:     []driver.Value{&ListParams{Page: 1, PerPage: 10, Date: "2020-01-01"}},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})

			// Exec insert wallets
			mock.
				ExpectQuery("select id, operation, wallet_from, wallet_to, amount, created_at from wallet_operations").
				WillReturnRows(rows)

		},
		expectedResultMatch: func(actual interface{}) bool {
			rows := actual.(chan *entities.WalletOperation)
			rowReceived := false
			for row := range rows {
				rowReceived = row.ID == 1
			}
			return rowReceived == false
		},
	},
	operationRepoTestCase{
		name:     "Failed receiving list of items (query row)",
		funcName: "List",
		args:     []driver.Value{nil},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})
			rows = rows.AddRow(1, Create, nil, 1, decimal.NewFromInt(0), time.Now())

			// Exec insert wallets
			mock.
				ExpectQuery("select id, operation, wallet_from, wallet_to, amount, created_at from wallet_operations").
				WillReturnError(fmt.Errorf("query error"))

		},
		err: fmt.Errorf("query error"),
	},
	operationRepoTestCase{
		name:     "Failed receiving list of items (scan row error)",
		funcName: "List",
		args:     []driver.Value{nil},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id", "operation", "wallet_from", "wallet_to", "amount", "created_at"})
			rows = rows.AddRow(nil, Create, nil, 1, decimal.NewFromInt(0), time.Now()).RowError(2, fmt.Errorf("scan error"))

			// Exec insert wallets
			mock.
				ExpectQuery("select id, operation, wallet_from, wallet_to, amount, created_at from wallet_operations").
				WillReturnRows(rows)

		},
		err: fmt.Errorf("[OPERATIONS_LIST_ROW]: sql: Scan error on column index 0, name \"id\": converting NULL to int is unsupported"),
	},
}

// Test operations repository actions
func TestOperationsRepo(t *testing.T) {
	for _, tc := range operationRepoTestCases {
		testLabel := strings.Join([]string{"Repo", "Operation", tc.name}, " ")
		t.Run(testLabel, func(t *testing.T) {
			ctx := context.Background()
			realArgs := []reflect.Value{
				reflect.ValueOf(ctx),
			}
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("cant create mock: %s", err)
			}
			defer db.Close()

			repo := WalletOperationService{
				db: db,
			}

			for _, arg := range tc.args {
				if arg == nil {
					pointer := (*ListParams)(nil)
					realArgs = append(realArgs, reflect.ValueOf(pointer))
				} else {
					realArgs = append(realArgs, reflect.ValueOf(arg))
				}
			}
			tc.mockQuery(mock)

			var result []reflect.Value
			if len(tc.args) > 0 {
				result = reflect.ValueOf(
					&repo,
				).MethodByName(
					tc.funcName,
				).Call(realArgs)
			} else {
				result = reflect.ValueOf(
					&repo,
				).MethodByName(
					tc.funcName,
				).Call(realArgs)
			}
			var reflectErr error
			resultValue := result[0].Interface()
			rerr := result[1].Interface()
			if rerr != nil {
				reflectErr = rerr.(error)
			}

			if reflectErr != nil && tc.err == nil {
				t.Errorf("unexpected err: %s", reflectErr)
				return
			}

			if tc.err != nil {
				if reflectErr == nil {
					t.Errorf("expected error, got nil: %s", reflectErr)
					return
				}
				if reflectErr != nil && !strings.Contains(reflectErr.Error(), tc.err.Error()) {
					t.Errorf("Expected string does not match receivede error. Got '%s'; Expected '%s'", reflectErr, tc.err)
					return
				}
			}

			if tc.err == nil && !tc.expectedResultMatch(resultValue) {
				t.Errorf("result data is not matched. Got %s", resultValue)
			}
		})
	}
}

// Test operation service constructor with transaction
func TestWithTransactionWalletOperationService(t *testing.T) {
	db, mock, _ := sqlmock.New()
	mock.ExpectBegin()
	txManager := tx.NewTxBeginner(db)
	localTx, _ := txManager.BeginTrx(context.Background(), nil)
	repo := NewWalletOperationRepo(db)
	repoWithTx := repo.WithTx(localTx)
	_, correctType := repoWithTx.(*WalletOperationService)
	if !correctType {
		t.Errorf("Wrong type of WalletOperationService")
	}
}

// Test service constructor
func TestNewWalletOperationService(t *testing.T) {
	db, _, _ := sqlmock.New()
	walletOperation := NewWalletOperationRepo(db)
	_, correctType := walletOperation.(*WalletOperationService)
	if !correctType {
		t.Errorf("Wrong type of WalletOperationService")
	}
}

// Benchmark repository Create operation
func BenchmarkCreate(b *testing.B) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"id"})
	rows = rows.AddRow(1)

	mock.
		ExpectQuery("insert into wallet_operations").
		WithArgs([]driver.Value{Create, 1, 2, decimal.NewFromInt(0)}...).
		WillReturnRows(rows)

	repo := NewWalletOperationRepo(sqlDB)
	for i := 0; i < b.N; i++ {
		_, _ = repo.Create(ctx, Create, 1, 2, decimal.NewFromInt(0))
	}
}

// Benchmark repository List operation
func BenchmarkList(b *testing.B) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()
	rows := sqlmock.NewRows([]string{"id"})
	rows = rows.AddRow(1)

	mock.
		ExpectQuery("select id, operation, wallet_from, wallet_to, amount, created_at from wallet_operations").
		WillReturnRows(rows)

	repo := NewWalletOperationRepo(sqlDB)

	for i := 0; i < b.N; i++ {
		_, _ = repo.List(ctx, nil)
	}
}
