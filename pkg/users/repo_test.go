package users

import (
	"billing_system_test_task/pkg/wallets"
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/shopspring/decimal"
)

type userRepoTestCase struct {
	name      string
	args      []driver.Value
	funcName  string
	mockQuery func(mock sqlmock.Sqlmock)
	err       error
}

var UserRepoTestCases = []userRepoTestCase{
	userRepoTestCase{
		name:     "Success user creation",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(mock sqlmock.Sqlmock) {
			userRows := sqlmock.NewRows([]string{"id"})
			userRows = userRows.AddRow(1)

			walletRows := sqlmock.NewRows([]string{"id"})
			walletRows = walletRows.AddRow(1)

			// Lock operation
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(CreateUser).
				WillReturnResult(sqlmock.NewResult(2, 2))

			// Start users create transaction
			mock.ExpectBegin()

			// Exec insert users query
			mock.
				ExpectQuery("insert into users").
				WithArgs([]driver.Value{"example@mail.com"}...).
				WillReturnRows(userRows)

			// Exec insert wallets query
			mock.
				ExpectQuery("insert into wallets").
				WithArgs([]driver.Value{int64(1)}...).
				WillReturnRows(walletRows)

			// Commit users create transaction
			mock.ExpectCommit()

			// Unlock operation
			mock.
				ExpectExec("select pg_advisory_unlock").
				WithArgs(CreateUser).
				WillReturnResult(sqlmock.NewResult(2, 2))
		},
	},
	userRepoTestCase{
		name:     "Failed user creation",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Lock operation
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(CreateUser).
				WillReturnResult(sqlmock.NewResult(2, 2))

			// Start users create transaction
			mock.ExpectBegin()

			// Exec insert users query with error
			mock.
				ExpectExec("insert into users").
				WithArgs([]driver.Value{"example@mail.com"}...).
				WillReturnError(fmt.Errorf("insert error"))

			// Rollback transaction
			mock.ExpectRollback()

			// Unlock operation
			mock.
				ExpectExec("select pg_advisory_unlock").
				WithArgs(CreateUser).
				WillReturnResult(sqlmock.NewResult(2, 2))
		},
		err: fmt.Errorf("Insert error"),
	},
	userRepoTestCase{
		name:     "Failed user creation (advisory lock error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Lock operation
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(CreateUser).
				WillReturnError(fmt.Errorf("Insert error (advisory_lock)"))
		},
		err: fmt.Errorf("Insert error (advisory_lock)"),
	},
	userRepoTestCase{
		name:     "Failed user creation (begin transaction error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Lock operation
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(CreateUser).
				WillReturnResult(sqlmock.NewResult(2, 2))

			// Begin users transaction with error
			mock.ExpectBegin().WillReturnError(fmt.Errorf("Errof of transaction start"))
		},
		err: fmt.Errorf("Insert error (begin transaction error)"),
	},
	userRepoTestCase{
		name:     "Failed user creation (wallets transaction commit error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(mock sqlmock.Sqlmock) {
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(CreateUser).
				WillReturnResult(sqlmock.NewResult(2, 2))

			// Start users create transaction
			mock.ExpectBegin()

			// Exec insert users query
			mock.
				ExpectExec("insert into users").
				WithArgs([]driver.Value{"example@mail.com"}...).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Start wallets create transaction
			mock.ExpectBegin()

			// Exec insert wallets query
			mock.
				ExpectExec("insert into wallets").
				WithArgs([]driver.Value{1}...).
				WillReturnResult(sqlmock.NewResult(1, 2))

			// Commit wallets create transaction
			mock.ExpectCommit().WillReturnError(fmt.Errorf("Transaction commit error"))
			// Commit users create transaction with error
			mock.ExpectRollback()
			// Unlock operation
			mock.
				ExpectExec("select pg_advisory_unlock").
				WithArgs(CreateUser).
				WillReturnResult(sqlmock.NewResult(2, 2))
		},
		err: fmt.Errorf("Insert error (wallet transaction error)"),
	},
	userRepoTestCase{
		name:     "Failed user creation (users transaction commit error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(mock sqlmock.Sqlmock) {
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(CreateUser).
				WillReturnResult(sqlmock.NewResult(2, 2))

			// Start users create transaction
			mock.ExpectBegin()

			// Exec insert users query
			mock.
				ExpectExec("insert into users").
				WithArgs([]driver.Value{"example@mail.com"}...).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Start wallets create transaction
			mock.ExpectBegin()

			// Exec insert wallets query
			mock.
				ExpectExec("insert into wallets").
				WithArgs([]driver.Value{1}...).
				WillReturnResult(sqlmock.NewResult(1, 2))

			// Commit wallets create transaction
			mock.ExpectCommit()
			// Commit users create transaction with error
			mock.ExpectCommit().WillReturnError(fmt.Errorf("Transaction commit error"))

			// Rollback users transaction
			mock.ExpectRollback()

			// Unlock operation
			mock.
				ExpectExec("select pg_advisory_unlock").
				WithArgs(CreateUser).
				WillReturnResult(sqlmock.NewResult(2, 2))
		},
		err: fmt.Errorf("Insert error (wallet transaction error)"),
	},
	userRepoTestCase{
		name:     "Failed user creation (advisory unlock error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Lock operation
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(CreateUser).
				WillReturnResult(sqlmock.NewResult(2, 2))

			// Start users create transaction
			mock.ExpectBegin()

			// Exec insert users query
			mock.
				ExpectExec("insert into users").
				WithArgs([]driver.Value{"example@mail.com"}...).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Start wallets create transaction
			mock.ExpectBegin()

			// Exec insert wallets query
			mock.
				ExpectExec("insert into wallets").
				WithArgs([]driver.Value{1}...).
				WillReturnResult(sqlmock.NewResult(1, 2))

			// Commit wallets create transaction
			mock.ExpectCommit()
			// Commit users create transaction
			mock.ExpectCommit()

			// Unlock operation
			mock.
				ExpectExec("select pg_advisory_unlock").
				WithArgs(CreateUser).
				WillReturnError(fmt.Errorf("advisory unlock error"))
		},
		err: fmt.Errorf("Insert error (advisor unlock error)"),
	},
	userRepoTestCase{
		name:     "Success user retrieving with id",
		funcName: "GetByID",
		args:     []driver.Value{1},
		mockQuery: func(mock sqlmock.Sqlmock) {
			query := `
				select u.id, u.email, w.balance, w.currency
			`
			rows := sqlmock.NewRows([]string{"id", "email", "balance", "currency"})
			rows = rows.AddRow(1, "test@example.com", decimal.NewFromInt(100), "USD")
			mock.
				ExpectQuery(query).
				WithArgs([]driver.Value{1}...).
				WillReturnRows(rows)
		},
	},
	userRepoTestCase{
		name:     "Failed user retrieving with id",
		funcName: "GetByID",
		args:     []driver.Value{1},
		mockQuery: func(mock sqlmock.Sqlmock) {
			query := `
				select u.id, u.email, w.balance, w.currency
				from users as u
				join wallets as w
				on u.id = w.user_id
				where u.id = $1
			`
			rows := sqlmock.NewRows([]string{"id", "email", "balance", "currency"})
			rows = rows.AddRow(nil, "test@example.com", decimal.NewFromInt(100), "USD")

			mock.
				ExpectQuery(query).
				WithArgs([]driver.Value{1}...).
				WillReturnError(fmt.Errorf("Row error"))
		},
		err: fmt.Errorf("Row error"),
	},
	userRepoTestCase{
		name:     "failed user retrieving with wallet ID (scan error)",
		funcName: "GetByID",
		args:     []driver.Value{1},
		mockQuery: func(mock sqlmock.Sqlmock) {
			query := `
				select u.id, u.email, w.balance, w.currency
				from users as u
				join wallets as w
				on u.id = w.user_id
				where u.id = $1
			`
			rows := sqlmock.NewRows([]string{"id", "email", "balance", "currency"}).
				AddRow(nil, "test@example.com", decimal.NewFromInt(100), "USD").
				RowError(1, fmt.Errorf("Scan error"))
			mock.
				ExpectQuery(query).
				WithArgs([]driver.Value{1}...).
				WillReturnRows(rows)
		},
		err: fmt.Errorf("Scan error"),
	},
	userRepoTestCase{
		name:     "Success user retrieving with wallet ID",
		funcName: "GetByWalletID",
		args:     []driver.Value{1},
		mockQuery: func(mock sqlmock.Sqlmock) {
			rows := sqlmock.NewRows([]string{"id", "email", "balance", "currency"})
			rows = rows.AddRow(1, "test@example.com", decimal.NewFromInt(100), "USD")
			mock.
				ExpectQuery("select u.id, u.email, w.balance, w.currency from users as u").
				WithArgs([]driver.Value{1}...).
				WillReturnRows(rows)
		},
	},
	userRepoTestCase{
		name:     "Failed user retrieving with wallet ID (sql error)",
		funcName: "GetByWalletID",
		args:     []driver.Value{1},
		mockQuery: func(mock sqlmock.Sqlmock) {
			mock.
				ExpectQuery("select u.id, u.email, w.balance, w.currency from users as u").
				WithArgs([]driver.Value{1}...).
				WillReturnError(fmt.Errorf("Error of user retrieving"))
		},
		err: fmt.Errorf("Error of user retrieving"),
	},
	userRepoTestCase{
		name:     "failed user retrieving with wallet ID (scan error)",
		funcName: "GetByWalletID",
		args:     []driver.Value{1},
		mockQuery: func(mock sqlmock.Sqlmock) {
			query := `
				select u.id, u.email, w.balance, w.currency
				from users as u
				join wallets as w
				on u.id = w.user_id
				where w.id = $1
			`
			rows := sqlmock.NewRows([]string{"id", "email", "balance", "currency"}).
				AddRow(nil, "test@example.com", decimal.NewFromInt(100), "USD").
				RowError(1, fmt.Errorf("Scan error"))
			mock.
				ExpectQuery(query).
				WithArgs([]driver.Value{1}...).
				WillReturnRows(rows)
		},
		err: fmt.Errorf("Scan error"),
	},
}

func TestUsersRepo(t *testing.T) {
	for _, tc := range UserRepoTestCases {
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

			walletsRepo := wallets.NewWalletService(db)
			repo := UsersService{
				db:          db,
				walletsRepo: walletsRepo,
			}

			for _, arg := range tc.args {
				realArgs = append(realArgs, reflect.ValueOf(arg))
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
			var (
				reflectErr error
			)
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

func TestNewUserService(t *testing.T) {
	db, _, _ := sqlmock.New()
	walletsRepo := wallets.NewWalletService(db)
	repo := NewUsersService(db, walletsRepo)
	_, correctType := repo.(UsersService)
	if !correctType {
		t.Errorf("Wrong type of UserService")
	}
}
