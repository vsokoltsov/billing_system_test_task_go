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

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
)

// userRepoTestCase saves information about user repo tests
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

			// Exec insert users query
			mock.
				ExpectQuery("insert into users").
				WithArgs([]driver.Value{"example@mail.com"}...).
				WillReturnRows(userRows)
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
			// Exec insert users query with error
			mock.
				ExpectExec("insert into users").
				WithArgs([]driver.Value{"example@mail.com"}...).
				WillReturnError(fmt.Errorf("insert error"))
		},
		err: fmt.Errorf("Insert error"),
	},
	userRepoTestCase{
		name:     "Failed user creation (scan error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(mock sqlmock.Sqlmock) {
			rows := sqlmock.NewRows([]string{"id", "email", "balance", "currency"}).
				AddRow(nil, "test@example.com", decimal.NewFromInt(100), "USD").
				RowError(1, fmt.Errorf("Scan error"))

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
			expectedUser := &entities.User{
				ID:    1,
				Email: "test@example.com",
				Wallet: &entities.Wallet{
					ID:       1,
					UserID:   1,
					Balance:  decimal.NewFromInt(100),
					Currency: "USD",
				},
			}
			actualUser := actual.(*entities.User)
			match := (expectedUser.ID == actualUser.ID &&
				expectedUser.Email == actualUser.Email &&
				expectedUser.Wallet.ID == actualUser.Wallet.ID &&
				expectedUser.Wallet.Balance.IntPart() == actualUser.Wallet.Balance.IntPart())
			return match
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
			expectedUser := &entities.User{
				ID:    1,
				Email: "test@example.com",
				Wallet: &entities.Wallet{
					ID:       1,
					UserID:   1,
					Balance:  decimal.NewFromInt(100),
					Currency: "USD",
				},
			}
			actualUser := actual.(*entities.User)
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

// Test user repository
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

			repo := UsersService{
				db: db,
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

			if tc.err == nil && !tc.expectedResultMatch(resultValue) {
				t.Errorf("result data is not matched. Got %s", resultValue)
			}
		})
	}
}

// Test user repository constructur
func TestNewUserService(t *testing.T) {
	db, _, _ := sqlmock.New()
	repo := NewUsersService(db)
	_, correctType := repo.(*UsersService)
	if !correctType {
		t.Errorf("Wrong type of UserService")
	}
}

func TestWithTransactionUserService(t *testing.T) {
	db, mock, _ := sqlmock.New()
	mock.ExpectBegin()
	txManager := tx.NewTxBeginner(db)
	localTx, _ := txManager.BeginTrx(context.Background(), nil)
	repo := NewUsersService(db)
	repoWithTx := repo.WithTx(localTx)
	_, correctType := repoWithTx.(*UsersService)
	if !correctType {
		t.Errorf("Wrong type of UserService")
	}
}

// Benchmarks repository's GetById operation
func BenchmarkGetById(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

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

	repo := NewUsersService(sqlDB)

	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByID(ctx, 1)
	}
}

// Benchmarks repository's GetByWalletID operation
func BenchmarkGetByWalletID(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

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

	repo := NewUsersService(sqlDB)
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByID(ctx, 1)
	}
}

// Benchmarks repository's Create operation
func BenchmarkCreateUser(b *testing.B) {
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

	repo := NewUsersService(sqlDB)

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
		WithArgs([]driver.Value{Create, nil, 1, amount}...).
		WillReturnRows(woRows)

	mock.ExpectCommit()

	mock.
		ExpectExec("select pg_advisory_unlock").
		WithArgs(CreateUser).
		WillReturnResult(sqlmock.NewResult(2, 2))

	for i := 0; i < b.N; i++ {
		_, _ = repo.Create(ctx, "example@mail.com")
	}
}
