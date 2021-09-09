package users

import (
	"billing_system_test_task/pkg/operations"
	"billing_system_test_task/pkg/wallets"
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
)

type userRepoTestCase struct {
	name                string
	args                []driver.Value
	funcName            string
	mockQuery           func(mock sqlmock.Sqlmock)
	err                 error
	expectedResultMatch func(actual interface{}) bool
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

			mock.
				ExpectQuery("insert into wallet_operations").
				WithArgs([]driver.Value{operations.Create, nil, 1, decimal.NewFromInt(0)}...).
				WillReturnRows(walletRows)

			// Commit users create transaction
			mock.ExpectCommit()

			// Unlock operation
			mock.
				ExpectExec("select pg_advisory_unlock").
				WithArgs(CreateUser).
				WillReturnResult(sqlmock.NewResult(2, 2))
		},
		expectedResultMatch: func(actual interface{}) bool {
			return actual.(int64) == int64(1)
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
		name:     "Failed user creation (wallet repo create error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(mock sqlmock.Sqlmock) {
			userRows := sqlmock.NewRows([]string{"id"})
			userRows = userRows.AddRow(1)

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
				WillReturnError(fmt.Errorf("User wallet creation error"))

			// Commit users create transaction with error
			mock.ExpectRollback()

			// Unlock operation
			mock.
				ExpectExec("select pg_advisory_unlock").
				WithArgs(CreateUser).
				WillReturnResult(sqlmock.NewResult(2, 2))
		},
		err: fmt.Errorf("User wallet creation error"),
	},
	userRepoTestCase{
		name:     "Failed user creation (users transaction commit error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(mock sqlmock.Sqlmock) {
			userRows := sqlmock.NewRows([]string{"id"})
			userRows = userRows.AddRow(1)

			walletRows := sqlmock.NewRows([]string{"id"})
			walletRows = walletRows.AddRow(1)

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

			// Exec insert wallets query
			mock.
				ExpectQuery("insert into wallet_operations").
				WithArgs([]driver.Value{int64(1)}...).
				WillReturnRows(walletRows)

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
		name:     "Failed user creation (scan error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(mock sqlmock.Sqlmock) {
			rows := sqlmock.NewRows([]string{"id", "email", "balance", "currency"}).
				AddRow(nil, "test@example.com", decimal.NewFromInt(100), "USD").
				RowError(1, fmt.Errorf("Scan error"))

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
				WillReturnRows(rows)
		},
		err: fmt.Errorf("Scan error"),
	},
	userRepoTestCase{
		name:     "Success user retrieving with id",
		funcName: "GetByID",
		args:     []driver.Value{1},
		mockQuery: func(mock sqlmock.Sqlmock) {
			query := "select u.id, u.email, w.id, w.user_id, w.balance, w.currency  from users as u"
			rows := sqlmock.NewRows([]string{"id", "email", "wallets.id", "user_id", "balance", "currency"})
			rows = rows.AddRow(1, "test@example.com", 1, 1, decimal.NewFromInt(100), "USD")
			mock.
				ExpectQuery(query).
				WithArgs([]driver.Value{1}...).
				WillReturnRows(rows)
		},
		expectedResultMatch: func(actual interface{}) bool {
			expectedUser := &User{
				ID:    1,
				Email: "test@example.com",
				Wallet: &wallets.Wallet{
					ID:       1,
					UserID:   1,
					Balance:  decimal.NewFromInt(100),
					Currency: "USD",
				},
			}
			actualUser := actual.(*User)
			return (expectedUser.ID == actualUser.ID &&
				expectedUser.Email == actualUser.Email &&
				expectedUser.Wallet.ID == actualUser.Wallet.ID &&
				expectedUser.Wallet.Balance.IntPart() == actualUser.Wallet.Balance.IntPart())
		},
	},
	userRepoTestCase{
		name:     "Failed user retrieving with id",
		funcName: "GetByID",
		args:     []driver.Value{1},
		mockQuery: func(mock sqlmock.Sqlmock) {
			query := `
				select u.id, u.email, w.id, w.user_id, w.balance, w.currency
			`

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
				select u.id, u.email, w.id, w.user_id, w.balance, w.currency
			`
			rows := sqlmock.NewRows([]string{"id", "email", "wallets.id", "user_id", "balance", "currency"}).
				AddRow(nil, "test@example.com", nil, 1, decimal.NewFromInt(100), "USD").
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
			rows := sqlmock.NewRows([]string{"id", "email", "wallets.id", "user_id", "balance", "currency"})
			rows = rows.AddRow(1, "test@example.com", 1, 1, decimal.NewFromInt(100), "USD")
			mock.
				ExpectQuery("select u.id, u.email, w.id, w.user_id, w.balance, w.currency from users as u").
				WithArgs([]driver.Value{1}...).
				WillReturnRows(rows)
		},
		expectedResultMatch: func(actual interface{}) bool {
			expectedUser := &User{
				ID:    1,
				Email: "test@example.com",
				Wallet: &wallets.Wallet{
					ID:       1,
					UserID:   1,
					Balance:  decimal.NewFromInt(100),
					Currency: "USD",
				},
			}
			actualUser := actual.(*User)
			return (expectedUser.ID == actualUser.ID &&
				expectedUser.Email == actualUser.Email &&
				expectedUser.Wallet.ID == actualUser.Wallet.ID &&
				expectedUser.Wallet.Balance.IntPart() == actualUser.Wallet.Balance.IntPart())
		},
	},
	userRepoTestCase{
		name:     "Failed user retrieving with wallet ID (sql error)",
		funcName: "GetByWalletID",
		args:     []driver.Value{1},
		mockQuery: func(mock sqlmock.Sqlmock) {
			mock.
				ExpectQuery("select u.id, u.email, w.id, w.user_id, w.balance, w.currency from users as u").
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
				select u.id, u.email, w.id, w.user_id, w.balance, w.currency
			`
			rows := sqlmock.NewRows([]string{"id", "email", "wallets.id", "user_id", "balance", "currency"}).
				AddRow(nil, "test@example.com", nil, 1, decimal.NewFromInt(100), "USD").
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

			walletOperation := operations.NewWalletOperationRepo(db)
			walletsRepo := wallets.NewWalletService(db, walletOperation)
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
			}

			// if resultValue != nil {
			if tc.err == nil && !tc.expectedResultMatch(resultValue) {
				t.Errorf("result data is not matched. Got %s", resultValue)
			}
			// }
		})
	}
}

func TestNewUserService(t *testing.T) {
	db, _, _ := sqlmock.New()
	walletOperation := operations.NewWalletOperationRepo(db)
	walletsRepo := wallets.NewWalletService(db, walletOperation)
	repo := NewUsersService(db, walletsRepo)
	_, correctType := repo.(UsersService)
	if !correctType {
		t.Errorf("Wrong type of UserService")
	}
}

func BenchmarkGetById(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	walletMock := wallets.NewMockWalletsManager(ctrl)
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()

	query := "select u.id, u.email, w.id, w.user_id, w.balance, w.currency  from users as u"
	rows := sqlmock.NewRows([]string{"id", "email", "wallets.id", "user_id", "balance", "currency"})
	rows = rows.AddRow(1, "test@example.com", 1, 1, decimal.NewFromInt(100), "USD")
	mock.
		ExpectQuery(query).
		WithArgs([]driver.Value{1}...).
		WillReturnRows(rows)

	repo := NewUsersService(sqlDB, walletMock)

	for i := 0; i < b.N; i++ {
		repo.GetByID(ctx, 1)
	}
}

func BenchmarkGetByWalletID(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	walletMock := wallets.NewMockWalletsManager(ctrl)
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "email", "wallets.id", "user_id", "balance", "currency"})
	rows = rows.AddRow(1, "test@example.com", 1, 1, decimal.NewFromInt(100), "USD")
	mock.
		ExpectQuery("select u.id, u.email, w.id, w.user_id, w.balance, w.currency from users as u").
		WithArgs([]driver.Value{1}...).
		WillReturnRows(rows)

	repo := NewUsersService(sqlDB, walletMock)
	for i := 0; i < b.N; i++ {
		repo.GetByID(ctx, 1)
	}
}

func BenchmarkCreate(b *testing.B) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()

	userRows := sqlmock.NewRows([]string{"id"})
	userRows = userRows.AddRow(1)

	walletRows := sqlmock.NewRows([]string{"id"})
	walletRows = walletRows.AddRow(1)

	woRows := sqlmock.NewRows([]string{"id"})
	woRows = woRows.AddRow(1)

	walletOperation := operations.NewWalletOperationRepo(sqlDB)
	walletsRepo := wallets.NewWalletService(sqlDB, walletOperation)
	repo := NewUsersService(sqlDB, walletsRepo)

	amount := decimal.NewFromInt(0)

	mock.
		ExpectExec("select pg_advisory_lock").
		WithArgs(CreateUser).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectBegin()

	mock.
		ExpectQuery("insert into users").
		WithArgs([]driver.Value{"example@mail.com"}...).
		WillReturnRows(userRows)

	mock.
		ExpectQuery("insert into wallets").
		WithArgs([]driver.Value{int64(1)}...).
		WillReturnRows(walletRows)

	mock.
		ExpectQuery("insert into wallet_operations").
		WithArgs([]driver.Value{operations.Create, nil, 1, amount}...).
		WillReturnRows(woRows)

	mock.ExpectCommit()

	mock.
		ExpectExec("select pg_advisory_unlock").
		WithArgs(CreateUser).
		WillReturnResult(sqlmock.NewResult(2, 2))

	for i := 0; i < b.N; i++ {
		repo.Create(ctx, "example@mail.com")
	}
}
