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
	"github.com/shopspring/decimal"
)

// walletRepoTestCase represents data for wallet repository test cases
type walletRepoTestCase struct {
	name                string
	queryMock           sqlQueryMock
	funcName            string
	mockQuery           func(mock sqlmock.Sqlmock)
	err                 error
	expectedResultMatch func(actual interface{}) bool
}

type sqlQueryMock struct {
	query       string
	args        []driver.Value
	columns     []string
	err         error
	requestType string
}

var WalletsRepoTestCase = []walletRepoTestCase{
	walletRepoTestCase{
		name:     "Success wallet creation",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:   "insert into wallets",
			args:    []driver.Value{int64(1)},
			columns: []string{"user_id"},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			rows := sqlmock.NewRows([]string{"id"})
			rows = rows.AddRow(1)

			// Exec insert wallets
			mock.
				ExpectQuery("insert into wallets").
				WithArgs([]driver.Value{int64(1)}...).
				WillReturnRows(rows)

		},
		expectedResultMatch: func(actual interface{}) bool {
			return actual.(int64) == int64(1)
		},
	},
	walletRepoTestCase{
		name:     "Failed wallet creation (insert error)",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query: "insert into wallets",
			args:  []driver.Value{int64(1)},
			err:   fmt.Errorf("Insert error"),
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Exec insert wallets
			mock.
				ExpectQuery("insert into wallets").
				WithArgs([]driver.Value{int64(1)}...).
				WillReturnError(fmt.Errorf("Insert error"))
		},
		err: fmt.Errorf("Insert error"),
	},
	walletRepoTestCase{
		name:     "Failed wallet creation (Scan error)",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:       "insert into wallets",
			args:        []driver.Value{int64(1)},
			columns:     []string{"user_id"},
			requestType: "insert-error",
			err:         fmt.Errorf("Insert error"),
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			rows := sqlmock.NewRows([]string{"id", "balance"})
			rows = rows.AddRow(nil, decimal.NewFromInt(0)).RowError(1, fmt.Errorf("Scan error"))

			// Exec insert wallets
			mock.
				ExpectQuery("insert into wallets").
				WithArgs([]driver.Value{int64(1)}...).
				WillReturnRows(rows)
		},
		err: fmt.Errorf("Scan error"),
	},
	walletRepoTestCase{
		name:     "Success wallet enroll",
		funcName: "Enroll",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{2, decimal.NewFromInt(100)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Exec update wallet balance query
			mock.
				ExpectExec("update wallets").
				WithArgs([]driver.Value{decimal.NewFromInt(100), 2}...).
				WillReturnResult(sqlmock.NewResult(1, 1))
		},
		expectedResultMatch: func(actual interface{}) bool {
			return actual.(int) == 2
		},
	},
	walletRepoTestCase{
		name:     "Failed wallet enroll (amount value is zero)",
		funcName: "Enroll",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{2, decimal.NewFromInt(0)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {},
		err:       fmt.Errorf("error of amount value"),
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

			mock.
				ExpectExec("update wallets").
				WithArgs([]driver.Value{decimal.NewFromInt(100), 2}...).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Exec update wallet balance query
			mock.
				ExpectExec("update wallets").
				WithArgs([]driver.Value{decimal.NewFromInt(100), 2}...).
				WillReturnError(fmt.Errorf("Update error (SQL update error)"))
		},
		err: fmt.Errorf("Update error (SQL update error)"),
	},
	walletRepoTestCase{
		name:     "Success wallet retrieving by user id",
		funcName: "GetByUserId",
		queryMock: sqlQueryMock{
			query: "select id, user_id, balance, currency from wallets",
			args:  []driver.Value{1},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			rows := sqlmock.NewRows([]string{"id", "user_id", "balance", "currency"})
			rows = rows.AddRow(1, 1, 100, "USD")
			mock.
				ExpectQuery("select id, user_id, balance, currency from wallets").
				WithArgs([]driver.Value{1}...).
				WillReturnRows(rows)
		},
		expectedResultMatch: func(actual interface{}) bool {
			actualWallet := actual.(*entities.Wallet)
			return actualWallet.ID == 1 && actualWallet.UserID == 1 && actualWallet.Balance.IntPart() == int64(100) && actualWallet.Currency == "USD"
		},
	},
	walletRepoTestCase{
		name:     "failed wallet retrieving by user id",
		funcName: "GetByUserId",
		queryMock: sqlQueryMock{
			query: "select id, user_id, balance, currency from wallets",
			args:  []driver.Value{1},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			mock.
				ExpectQuery("select id, user_id, balance, currency from wallets").
				WithArgs([]driver.Value{1}...).
				WillReturnError(fmt.Errorf("Wallet retrieving error"))
		},
		err: fmt.Errorf("Wallet retrieving error"),
	},
	walletRepoTestCase{
		name:     "Success wallet transfer funds",
		funcName: "Transfer",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{1, 2, decimal.NewFromInt(10)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {

			// Select source wallet
			rows := sqlmock.NewRows([]string{"id", "user_id", "balance", "currency"})
			rows = rows.AddRow(1, 1, 100, "USD")
			mock.
				ExpectQuery("select id, user_id, balance, currency from wallets").
				WithArgs([]driver.Value{1}...).
				WillReturnRows(rows)

			// Exec update wallet balance query
			mock.
				ExpectExec("update wallets").
				WithArgs([]driver.Value{decimal.NewFromInt(10), 1}...).
				WillReturnResult(sqlmock.NewResult(0, 1))

			// Exec update wallet balance query
			mock.
				ExpectExec("update wallets").
				WithArgs([]driver.Value{decimal.NewFromInt(10), 2}...).
				WillReturnResult(sqlmock.NewResult(0, 1))
		},
		expectedResultMatch: func(actual interface{}) bool {
			actualID := actual.(int)
			return actualID == 1
		},
	},
	walletRepoTestCase{
		name:     "Failed wallet transfer funds (source wallet balance is zero)",
		funcName: "Transfer",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{1, 2, decimal.NewFromInt(10)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Select source wallet
			rows := sqlmock.NewRows([]string{"id", "user_id", "balance", "currency"})
			rows = rows.AddRow(1, 1, 0, "USD")
			mock.
				ExpectQuery("select id, user_id, balance, currency from wallets").
				WithArgs([]driver.Value{1}...).
				WillReturnRows(rows)
		},
		err: fmt.Errorf("Source wallet balance is equal to zero"),
	},
	walletRepoTestCase{
		name:     "Failed wallet transfer funds (select source wallet error)",
		funcName: "Transfer",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{1, 2, decimal.NewFromInt(10)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Select source wallet error
			mock.
				ExpectQuery("select id, user_id, balance, currency from wallets").
				WithArgs([]driver.Value{1}...).
				WillReturnError(fmt.Errorf("error of receiving source wallet"))
		},
		err: fmt.Errorf("error of receiving source wallet"),
	},
	walletRepoTestCase{
		name:     "Failed wallet transfer funds (update source wallet error)",
		funcName: "Transfer",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{1, 2, decimal.NewFromInt(10)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Select source wallet
			rows := sqlmock.NewRows([]string{"id", "user_id", "balance", "currency"})
			rows = rows.AddRow(1, 1, 100, "USD")
			mock.
				ExpectQuery("select id, user_id, balance, currency from wallets").
				WithArgs([]driver.Value{1}...).
				WillReturnRows(rows)

			// Lock operation
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(TransferFunds).
				WillReturnResult(sqlmock.NewResult(2, 2))

			// Start transfer funds transaction
			mock.ExpectBegin()

			// Exec update wallet balance query
			mock.
				ExpectExec("update wallets").
				WithArgs([]driver.Value{decimal.NewFromInt(10), 1}...).
				WillReturnError(fmt.Errorf("update source wallet error"))
		},
		err: fmt.Errorf("update source wallet error"),
	},
	walletRepoTestCase{
		name:     "Failed wallet transfer funds (update target wallet error)",
		funcName: "Transfer",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{1, 2, decimal.NewFromInt(10)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Select source wallet
			rows := sqlmock.NewRows([]string{"id", "user_id", "balance", "currency"})
			rows = rows.AddRow(1, 1, 100, "USD")
			mock.
				ExpectQuery("select id, user_id, balance, currency from wallets").
				WithArgs([]driver.Value{1}...).
				WillReturnRows(rows)

			// Lock operation
			mock.
				ExpectExec("select pg_advisory_lock").
				WithArgs(TransferFunds).
				WillReturnResult(sqlmock.NewResult(2, 2))

			// Start transfer funds transaction
			mock.ExpectBegin()

			// Exec update wallet balance query
			mock.
				ExpectExec("update wallets").
				WithArgs([]driver.Value{decimal.NewFromInt(10), 1}...).
				WillReturnResult(sqlmock.NewResult(0, 1))

			// Exec update wallet balance query
			mock.
				ExpectExec("update wallets").
				WithArgs([]driver.Value{decimal.NewFromInt(10), 2}...).
				WillReturnError(fmt.Errorf("update target wallet error"))
		},
		err: fmt.Errorf("update target wallet error"),
	},
	walletRepoTestCase{
		name:     "Failed wallet transfer funds (update receiver wallet error)",
		funcName: "Transfer",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{1, 2, decimal.NewFromInt(10)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Select source wallet
			rows := sqlmock.NewRows([]string{"id", "user_id", "balance", "currency"})
			rows = rows.AddRow(1, 1, 100, "USD")
			mock.
				ExpectQuery("select id, user_id, balance, currency from wallets").
				WithArgs([]driver.Value{1}...).
				WillReturnRows(rows)

			// Exec update wallet balance query
			mock.
				ExpectExec("update wallets").
				WithArgs([]driver.Value{decimal.NewFromInt(10), 1}...).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Exec update wallet balance query
			mock.
				ExpectExec("update wallets").
				WithArgs([]driver.Value{decimal.NewFromInt(10), 2}...).
				WillReturnError(fmt.Errorf("update target wallet error"))

		},
		err: fmt.Errorf("update target wallet error"),
	},
}

// Tests wallets repository
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

// Tests wallets repository constructor
func TestNewWalletService(t *testing.T) {
	db, _, _ := sqlmock.New()
	walletsRepo := NewWalletService(db)
	_, correctType := walletsRepo.(*WalletService)
	if !correctType {
		t.Errorf("Wrong type of WalletOperations")
	}
}

func TestWithTransactionWalletService(t *testing.T) {
	db, mock, _ := sqlmock.New()
	mock.ExpectBegin()
	txManager := tx.NewTxBeginner(db)
	localTx, _ := txManager.BeginTrx(context.Background(), nil)
	repo := NewWalletService(db)
	repoWithTx := repo.WithTx(localTx)
	_, correctType := repoWithTx.(*WalletService)
	if !correctType {
		t.Errorf("Wrong type of WalletService")
	}
}

// Tests repository Create action
func BenchmarkCreateWallet(b *testing.B) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()

	mock.ExpectBegin()

	walletRows := sqlmock.NewRows([]string{"id"})
	walletRows = walletRows.AddRow(1)

	woRows := sqlmock.NewRows([]string{"id"})
	woRows = woRows.AddRow(1)

	repo := NewWalletService(sqlDB)

	mock.
		ExpectQuery("insert into wallets").
		WithArgs([]driver.Value{int64(1)}...).
		WillReturnRows(walletRows)

	for i := 0; i < b.N; i++ {
		_, _ = repo.Create(ctx, int64(1))
	}
}

// Tests repository Enroll action
func BenchmarkEnroll(b *testing.B) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()

	repo := NewWalletService(sqlDB)

	mock.
		ExpectExec("update wallets").
		WithArgs([]driver.Value{decimal.NewFromInt(100), 2}...).
		WillReturnResult(sqlmock.NewResult(1, 1))

	for i := 0; i < b.N; i++ {
		_, _ = repo.Enroll(ctx, 1, decimal.NewFromInt(100))
	}
}

// Tests repository GetByUserID action
func BenchmarkGetByUserID(b *testing.B) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "user_id", "balance", "currency"})
	rows = rows.AddRow(1, 1, 100, "USD")

	repo := NewWalletService(sqlDB)

	mock.
		ExpectQuery("select id, user_id, balance, currency from wallets").
		WithArgs([]driver.Value{1}...).
		WillReturnRows(rows)

	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByUserId(ctx, 1)
	}
}

// Tests repository GetByID action
func BenchmarkGetByID(b *testing.B) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "user_id", "balance", "currency"})
	rows = rows.AddRow(1, 1, 100, "USD")

	// walletOperation := NewWalletOperationRepo(sqlDB)
	repo := NewWalletService(sqlDB)

	mock.
		ExpectQuery("select id, user_id, balance, currency from wallets").
		WithArgs([]driver.Value{1}...).
		WillReturnRows(rows)

	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByID(ctx, 1)
	}
}

// Tests repository Transfer action
func BenchmarkTransfer(b *testing.B) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()

	repo := NewWalletService(sqlDB)

	// Select source wallet
	rows := sqlmock.NewRows([]string{"id", "user_id", "balance", "currency"})
	rows = rows.AddRow(1, 1, 100, "USD")
	mock.
		ExpectQuery("select id, user_id, balance, currency from wallets").
		WithArgs([]driver.Value{1}...).
		WillReturnRows(rows)

	// Start transfer funds transaction
	mock.ExpectBegin()

	// Exec update wallet balance query
	mock.
		ExpectExec("update wallets").
		WithArgs([]driver.Value{decimal.NewFromInt(10), 1}...).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Exec update wallet balance query
	mock.
		ExpectExec("update wallets").
		WithArgs([]driver.Value{decimal.NewFromInt(10), 2}...).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Unlock operation
	mock.
		ExpectExec("select pg_advisory_unlock").
		WithArgs(TransferFunds).
		WillReturnResult(sqlmock.NewResult(2, 2))

	for i := 0; i < b.N; i++ {
		_, _ = repo.Transfer(ctx, 1, 2, decimal.NewFromInt(10))
	}
}
