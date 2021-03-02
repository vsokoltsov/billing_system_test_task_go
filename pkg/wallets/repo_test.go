package wallets

import (
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type walletRepoTestCase struct {
	name      string
	queryMock sqlQueryMock
	funcName  string
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
		name:     "Success user creation",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:       "insert into wallets",
			args:        []driver.Value{1},
			columns:     []string{"user_id"},
			requestType: "insert",
			mockResult:  sqlmock.NewResult(1, 1),
			mockQuery: func(mock sqlmock.Sqlmock, query string, args []driver.Value, result driver.Result, rows *sqlmock.Rows, err error) {
				mock.
					ExpectExec("select pg_advisory_lock").
					WithArgs(CreateWallet).
					WillReturnResult(sqlmock.NewResult(2, 2))

				mock.ExpectBegin()

				mock.
					ExpectExec(query).
					WithArgs(args...).
					WillReturnResult(result)

				mock.ExpectCommit()

				mock.
					ExpectExec("select pg_advisory_unlock").
					WithArgs(CreateWallet).
					WillReturnResult(sqlmock.NewResult(2, 2))
			},
		},
	},
	walletRepoTestCase{
		name:     "Failed user creation",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:       "insert into wallets",
			args:        []driver.Value{1},
			columns:     []string{"user_id"},
			requestType: "insert-error",
			err:         fmt.Errorf("Insert error"),
			mockQuery: func(mock sqlmock.Sqlmock, query string, args []driver.Value, result driver.Result, rows *sqlmock.Rows, err error) {

				mock.
					ExpectExec("select pg_advisory_lock").
					WithArgs(CreateWallet).
					WillReturnResult(sqlmock.NewResult(2, 2))

				mock.ExpectBegin()

				mock.
					ExpectExec(query).
					WithArgs(args...).
					WillReturnError(err)

				mock.ExpectRollback()
				mock.
					ExpectExec("select pg_advisory_unlock").
					WithArgs(CreateWallet).
					WillReturnResult(sqlmock.NewResult(2, 2))
			},
		},
	},
	walletRepoTestCase{
		name:     "Failed user creation (advisory lock error)",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:       "insert into wallets",
			args:        []driver.Value{1},
			columns:     []string{"user_id"},
			requestType: "insert-error",
			err:         fmt.Errorf("Insert error (advisory_lock)"),
			mockQuery: func(mock sqlmock.Sqlmock, query string, args []driver.Value, result driver.Result, rows *sqlmock.Rows, err error) {
				mock.
					ExpectExec("select pg_advisory_lock").
					WithArgs(CreateWallet).
					WillReturnError(fmt.Errorf("Insert error (advisory_lock)"))
			},
		},
	},
	walletRepoTestCase{
		name:     "Failed user creation (begin transaction error)",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:       "insert into wallets",
			args:        []driver.Value{1},
			columns:     []string{"user_id"},
			requestType: "insert-error",
			err:         fmt.Errorf("Insert error (begin transaction error)"),
			mockQuery: func(mock sqlmock.Sqlmock, query string, args []driver.Value, result driver.Result, rows *sqlmock.Rows, err error) {

				mock.
					ExpectExec("select pg_advisory_lock").
					WithArgs(CreateWallet).
					WillReturnResult(sqlmock.NewResult(2, 2))

				mock.ExpectBegin().WillReturnError(fmt.Errorf("Errof of transaction start"))
			},
		},
	},
	walletRepoTestCase{
		name:     "Failed user creation (transaction commit error)",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:       "insert into wallets",
			args:        []driver.Value{1},
			columns:     []string{"user_id"},
			requestType: "insert-error",
			err:         fmt.Errorf("Insert error (transaction commit error)"),
			mockResult:  sqlmock.NewResult(1, 1),
			mockQuery: func(mock sqlmock.Sqlmock, query string, args []driver.Value, result driver.Result, rows *sqlmock.Rows, err error) {

				mock.
					ExpectExec("select pg_advisory_lock").
					WithArgs(CreateWallet).
					WillReturnResult(sqlmock.NewResult(2, 2))

				mock.ExpectBegin()

				mock.
					ExpectExec(query).
					WithArgs(args...).
					WillReturnResult(result)

				mock.ExpectCommit().WillReturnError(fmt.Errorf("Transaction commit error"))
			},
		},
	},
	walletRepoTestCase{
		name:     "Failed user creation (advisory unlock error)",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:       "insert into wallets",
			args:        []driver.Value{1},
			columns:     []string{"user_id"},
			requestType: "insert-error",
			mockResult:  sqlmock.NewResult(1, 1),
			err:         fmt.Errorf("Insert error (advisor unlock error)"),
			mockQuery: func(mock sqlmock.Sqlmock, query string, args []driver.Value, result driver.Result, rows *sqlmock.Rows, err error) {

				mock.
					ExpectExec("select pg_advisory_lock").
					WithArgs(CreateWallet).
					WillReturnResult(sqlmock.NewResult(2, 2))

				mock.ExpectBegin()

				mock.
					ExpectExec(query).
					WithArgs(args...).
					WillReturnResult(result)

				mock.ExpectCommit()

				mock.
					ExpectExec("select pg_advisory_unlock").
					WithArgs(CreateWallet).
					WillReturnError(fmt.Errorf("advisory unlock error"))
			},
		},
	},
}

func TestWalletRepo(t *testing.T) {
	for _, tc := range WalletsRepoTestCase {
		testLabel := strings.Join([]string{"Repo", "User", tc.name}, " ")
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

			repo := WalletService{
				db: db,
			}

			rows := sqlmock.
				NewRows(tc.queryMock.columns)

			for _, row := range tc.queryMock.values {
				if tc.queryMock.err != nil {
					rows = rows.AddRow(nil, row["email"]).RowError(row["id"].(int), tc.queryMock.err)
				} else {
					rows = rows.AddRow(row["id"], row["email"])
				}
			}

			for _, arg := range tc.queryMock.args {
				realArgs = append(realArgs, reflect.ValueOf(arg))
			}
			tc.queryMock.mockQuery(mock, tc.queryMock.query, tc.queryMock.args, tc.queryMock.mockResult, rows, tc.queryMock.err)

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

			if reflectErr != nil && tc.queryMock.err == nil {
				t.Errorf("unexpected err: %s", reflectErr)
				return
			}

			if tc.queryMock.err != nil {
				if reflectErr == nil {
					t.Errorf("expected error, got nil: %s", reflectErr)
					return
				}
			}
		})
	}
}
