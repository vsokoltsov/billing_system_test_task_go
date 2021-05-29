package wallets

import (
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
)

type walletRepoTestCase struct {
	name      string
	queryMock sqlQueryMock
	funcName  string
	mockQuery func(mock sqlmock.Sqlmock)
	err       error
}

type sqlQueryMock struct {
	query       string
	args        []driver.Value
	columns     []string
	values      []map[string]interface{}
	err         error
	requestType string
	mockResult  driver.Result
	mockQuery   func(mock sqlmock.Sqlmock, query string, args []driver.Value, result driver.Result, rows *sqlmock.Rows, err error)
}

var WalletsRepoTestCase = []walletRepoTestCase{
	walletRepoTestCase{
		name:     "Success wallet creation",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:       "insert into wallets",
			args:        []driver.Value{int64(1)},
			columns:     []string{"user_id"},
			requestType: "insert",
			mockResult:  sqlmock.NewResult(1, 1),
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			mock.ExpectBegin()

			// Exec insert wallets
			mock.
				ExpectExec("insert into wallets").
				WithArgs([]driver.Value{int64(1)}...).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Commit transaction
			mock.ExpectCommit()
		},
	},
	walletRepoTestCase{
		name:     "Failed wallet creation (insert error)",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:       "insert into wallets",
			args:        []driver.Value{int64(1)},
			columns:     []string{"user_id"},
			requestType: "insert-error",
			err:         fmt.Errorf("Insert error"),
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			mock.ExpectBegin()

			// Exec insert wallets
			mock.
				ExpectExec("insert into wallets").
				WithArgs([]driver.Value{int64(1)}...).
				WillReturnError(fmt.Errorf("Insert error"))

			// Rollback transaction
			mock.ExpectRollback()
		},
		err: fmt.Errorf("Insert error"),
	},
	walletRepoTestCase{
		name:     "Failed wallet creation (transaction begin error)",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:       "insert into wallets",
			args:        []driver.Value{int64(1)},
			columns:     []string{"user_id"},
			requestType: "insert-error",
			err:         fmt.Errorf("Insert error"),
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			mock.ExpectBegin().WillReturnError(fmt.Errorf("Begin transaction error"))
		},
		err: fmt.Errorf("Begin error"),
	},
	walletRepoTestCase{
		name:     "Failed wallet creation (transaction commit error)",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:       "insert into wallets",
			args:        []driver.Value{int64(1)},
			columns:     []string{"user_id"},
			requestType: "insert-error",
			err:         fmt.Errorf("Insert error"),
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			mock.ExpectBegin()

			// Exec insert wallets
			mock.
				ExpectExec("insert into wallets").
				WithArgs([]driver.Value{int64(1)}...).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Commit transaction with error
			mock.ExpectCommit().WillReturnError(fmt.Errorf("Commit error"))
		},
		err: fmt.Errorf("Commit error"),
	},
	walletRepoTestCase{
		name:     "Success wallet enroll",
		funcName: "Enroll",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{2, decimal.NewFromInt(100)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Lock operation
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(EnrollWallet).
				WillReturnResult(sqlmock.NewResult(2, 2))

			// Start wallet enroll transaction
			mock.ExpectBegin()

			// Exec update wallet balance query
			mock.
				ExpectExec("update wallets").
				WithArgs([]driver.Value{decimal.NewFromInt(100), 2}...).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Commit update wallet transaction
			mock.ExpectCommit()

			// Unlock operation
			mock.
				ExpectExec("select pg_advisory_unlock").
				WithArgs(EnrollWallet).
				WillReturnResult(sqlmock.NewResult(2, 2))
		},
	},
	walletRepoTestCase{
		name:     "Failed wallet enroll (advisory lock error)",
		funcName: "Enroll",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{2, decimal.NewFromInt(100)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Lock operation
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(EnrollWallet).
				WillReturnError(fmt.Errorf("Update error (advisory_lock)"))
		},
		err: fmt.Errorf("Update error (advisory_lock)"),
	},
	walletRepoTestCase{
		name:     "Failed wallet enroll (begin transaction error)",
		funcName: "Enroll",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{2, decimal.NewFromInt(100)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Lock operation
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(EnrollWallet).
				WillReturnResult(sqlmock.NewResult(2, 2))

			// Start wallet enroll transaction
			mock.ExpectBegin().WillReturnError(fmt.Errorf("Update error (Begin transaction error)"))
		},
		err: fmt.Errorf("Update error (Begin transaction error)"),
	},
	walletRepoTestCase{
		name:     "Failed wallet enroll (update query error)",
		funcName: "Enroll",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{2, decimal.NewFromInt(100)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Lock operation
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(EnrollWallet).
				WillReturnResult(sqlmock.NewResult(2, 2))

			// Start wallet enroll transaction
			mock.ExpectBegin()

			// Exec update wallet balance query
			mock.
				ExpectExec("update wallets").
				WithArgs([]driver.Value{decimal.NewFromInt(100), 2}...).
				WillReturnError(fmt.Errorf("Update error (SQL update error)"))
		},
		err: fmt.Errorf("Update error (SQL update error)"),
	},
	walletRepoTestCase{
		name:     "Failed wallet enroll (transaction commit error)",
		funcName: "Enroll",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{2, decimal.NewFromInt(100)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Lock operation
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(EnrollWallet).
				WillReturnResult(sqlmock.NewResult(2, 2))

			// Start wallet enroll transaction
			mock.ExpectBegin()

			// Exec update wallet balance query
			mock.
				ExpectExec("update wallets").
				WithArgs([]driver.Value{decimal.NewFromInt(100), 2}...).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Commit update wallet transaction
			mock.ExpectCommit().WillReturnError(fmt.Errorf("Update error (Trasnaction commit error)"))
		},
		err: fmt.Errorf("Update error (Trasnaction commit error)"),
	},
	walletRepoTestCase{
		name:     "Failed wallet enroll (advisory unlock error)",
		funcName: "Enroll",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{2, decimal.NewFromInt(100)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Lock operation
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(EnrollWallet).
				WillReturnResult(sqlmock.NewResult(2, 2))

			// Start wallet enroll transaction
			mock.ExpectBegin()

			// Exec update wallet balance query
			mock.
				ExpectExec("update wallets").
				WithArgs([]driver.Value{decimal.NewFromInt(100), 2}...).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Commit update wallet transaction
			mock.ExpectCommit()

			// Unlock operation
			mock.
				ExpectExec("select pg_advisory_unlock").
				WithArgs(EnrollWallet).
				WillReturnError(fmt.Errorf("Update error (Advisory unlock error)"))
		},
		err: fmt.Errorf("Update error (Advisory unlock error)"),
	},
}

func TestWalletRepo(t *testing.T) {
	for _, tc := range WalletsRepoTestCase {
		testLabel := strings.Join([]string{"Repo", "Wallet", tc.name}, " ")
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

			// Create Connection instance for Create wallet tests
			if tc.funcName == "Create" {
				conn, _ := db.Conn(ctx)
				realArgs = append(realArgs, reflect.ValueOf(conn))
			}

			repo := WalletService{
				db: db,
			}

			for _, arg := range tc.queryMock.args {
				realArgs = append(realArgs, reflect.ValueOf(arg))
			}
			tc.mockQuery(mock)

			var result []reflect.Value
			if len(tc.queryMock.args) > 0 {
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
			}
		})
	}
}

func TestNewWalletService(t *testing.T) {
	db, _, _ := sqlmock.New()
	walletsRepo := NewWalletService(db)
	_, correctType := walletsRepo.(WalletService)
	if !correctType {
		t.Errorf("Wrong type of UserService")
	}
}
