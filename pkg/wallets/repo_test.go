package wallets

import (
	"billing_system_test_task/pkg/operations"
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
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id"})
			rows = rows.AddRow(1)

			operationRows := sqlmock.NewRows([]string{"id"}).AddRow(1)

			// Exec insert wallets
			mock.
				ExpectQuery("insert into wallets").
				WithArgs([]driver.Value{int64(1)}...).
				WillReturnRows(rows)

			mock.
				ExpectQuery("insert into wallet_operations").
				WithArgs([]driver.Value{operations.Create, nil, 1, decimal.NewFromInt(0)}...).
				WillReturnRows(operationRows)

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
		name:     "Failed wallet creation (Operation create error)",
		funcName: "Create",
		queryMock: sqlQueryMock{
			query:   "insert into wallets",
			args:    []driver.Value{int64(1)},
			columns: []string{"user_id"},
			err:     fmt.Errorf("Insert error"),
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id"})
			rows = rows.AddRow(1)

			// Exec insert wallets
			mock.
				ExpectQuery("insert into wallets").
				WithArgs([]driver.Value{int64(1)}...).
				WillReturnRows(rows)

			mock.
				ExpectQuery("insert into wallet_operations").
				WithArgs([]driver.Value{operations.Create, nil, 1, decimal.NewFromInt(0)}...).
				WillReturnError(fmt.Errorf("Operation error"))
		},
		expectedResultMatch: func(actual interface{}) bool {
			return actual.(int64) == int64(1)
		},
	},
	walletRepoTestCase{
		name:     "Success wallet enroll",
		funcName: "Enroll",
		queryMock: sqlQueryMock{
			query: "update wallets set",
			args:  []driver.Value{2, decimal.NewFromInt(100)},
		},
		mockQuery: func(mock sqlmock.Sqlmock) {
			operationRows := sqlmock.NewRows([]string{"id"}).AddRow(1)

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

			mock.
				ExpectQuery("insert into wallet_operations").
				WithArgs([]driver.Value{operations.Deposit, nil, 2, decimal.NewFromInt(100)}...).
				WillReturnRows(operationRows)

			// Commit update wallet transaction
			mock.ExpectCommit()

			// Unlock operation
			mock.
				ExpectExec("select pg_advisory_unlock").
				WithArgs(EnrollWallet).
				WillReturnResult(sqlmock.NewResult(2, 2))
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
			actualWallet := actual.(*Wallet)
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
			sourceOperationRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
			destinationOperationRows := sqlmock.NewRows([]string{"id"}).AddRow(2)

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
				WillReturnResult(sqlmock.NewResult(0, 1))

			mock.
				ExpectQuery("insert into wallet_operations").
				WithArgs([]driver.Value{operations.Deposit, 1, 2, decimal.NewFromInt(10)}...).
				WillReturnRows(sourceOperationRows)

			mock.
				ExpectQuery("insert into wallet_operations").
				WithArgs([]driver.Value{operations.Withdrawal, 2, 1, decimal.NewFromInt(10)}...).
				WillReturnRows(destinationOperationRows)

			// Commit transfer funds transaction
			mock.ExpectCommit()

			// Unlock operation
			mock.
				ExpectExec("select pg_advisory_unlock").
				WithArgs(TransferFunds).
				WillReturnResult(sqlmock.NewResult(2, 2))
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
		name:     "Failed wallet transfer funds (advisory lock error)",
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
				WillReturnError(fmt.Errorf("advisory lock error"))
		},
		err: fmt.Errorf("advisory Lock error"),
	},
	walletRepoTestCase{
		name:     "Failed wallet transfer funds (transaction beg",
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
			mock.ExpectBegin().WillReturnError(fmt.Errorf("transaction begin errror"))
		},
		err: fmt.Errorf("transaction begin errror"),
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
		name:     "Failed wallet transfer funds (transaction commit error)",
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
				WillReturnResult(sqlmock.NewResult(0, 1))

			// Commit transfer funds transaction
			mock.ExpectCommit().WillReturnError(fmt.Errorf("transaction commit error"))
		},
		err: fmt.Errorf("transaction commit error"),
	},
	walletRepoTestCase{
		name:     "Failed wallet transfer funds (advisory unlock error)",
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
				WillReturnResult(sqlmock.NewResult(0, 1))

			// Commit transfer funds transaction
			mock.ExpectCommit()

			// Unlock operation
			mock.
				ExpectExec("select pg_advisory_unlock").
				WithArgs(TransferFunds).
				WillReturnError(fmt.Errorf("advisory unlock error"))
		},
		err: fmt.Errorf("advisory unlock error"),
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
				mock.ExpectBegin()
				conn, _ := db.Conn(ctx)
				tx, txErr := conn.BeginTx(ctx, nil)
				if txErr != nil {
					fmt.Printf("error of transaction initialization: %s", txErr)
				}
				realArgs = append(realArgs, reflect.ValueOf(tx))
			}

			wos := operations.NewWalletOperationRepo(db)
			repo := WalletService{
				db:                  db,
				walletOperationRepo: wos,
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

func TestNewWalletService(t *testing.T) {
	db, _, _ := sqlmock.New()
	walletOperation := operations.NewWalletOperationRepo(db)
	walletsRepo := NewWalletService(db, walletOperation)
	_, correctType := walletsRepo.(WalletService)
	if !correctType {
		t.Errorf("Wrong type of WalletOperations")
	}
}

func BenchmarkCreate(b *testing.B) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()

	mock.ExpectBegin()
	conn, _ := sqlDB.Conn(ctx)
	tx, txErr := conn.BeginTx(ctx, nil)
	if txErr != nil {
		fmt.Printf("error of transaction initialization: %s", txErr)
	}

	walletRows := sqlmock.NewRows([]string{"id"})
	walletRows = walletRows.AddRow(1)

	woRows := sqlmock.NewRows([]string{"id"})
	woRows = woRows.AddRow(1)

	walletOperation := operations.NewWalletOperationRepo(sqlDB)
	repo := NewWalletService(sqlDB, walletOperation)

	mock.
		ExpectQuery("insert into wallets").
		WithArgs([]driver.Value{int64(1)}...).
		WillReturnRows(walletRows)

	mock.
		ExpectQuery("insert into wallet_operations").
		WithArgs([]driver.Value{operations.Create, nil, 1, decimal.NewFromInt(0)}...).
		WillReturnRows(woRows)

	for i := 0; i < b.N; i++ {
		repo.Create(ctx, tx, int64(1))
	}
}

func BenchmarkEnroll(b *testing.B) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()

	walletOperation := operations.NewWalletOperationRepo(sqlDB)
	repo := NewWalletService(sqlDB, walletOperation)

	operationRows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.
		ExpectExec("select pg_advisory_lock").
		WithArgs(EnrollWallet).
		WillReturnResult(sqlmock.NewResult(2, 2))

	mock.ExpectBegin()

	mock.
		ExpectExec("update wallets").
		WithArgs([]driver.Value{decimal.NewFromInt(100), 2}...).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.
		ExpectQuery("insert into wallet_operations").
		WithArgs([]driver.Value{operations.Deposit, nil, 2, decimal.NewFromInt(100)}...).
		WillReturnRows(operationRows)

	mock.ExpectCommit()

	mock.
		ExpectExec("select pg_advisory_unlock").
		WithArgs(EnrollWallet).
		WillReturnResult(sqlmock.NewResult(2, 2))

	for i := 0; i < b.N; i++ {
		repo.Enroll(ctx, 1, decimal.NewFromInt(100))
	}
}

func BenchmarkGetByUserID(b *testing.B) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "user_id", "balance", "currency"})
	rows = rows.AddRow(1, 1, 100, "USD")

	walletOperation := operations.NewWalletOperationRepo(sqlDB)
	repo := NewWalletService(sqlDB, walletOperation)

	mock.
		ExpectQuery("select id, user_id, balance, currency from wallets").
		WithArgs([]driver.Value{1}...).
		WillReturnRows(rows)

	for i := 0; i < b.N; i++ {
		repo.GetByUserId(ctx, 1)
	}
}

func BenchmarkGetByID(b *testing.B) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "user_id", "balance", "currency"})
	rows = rows.AddRow(1, 1, 100, "USD")

	walletOperation := operations.NewWalletOperationRepo(sqlDB)
	repo := NewWalletService(sqlDB, walletOperation)

	mock.
		ExpectQuery("select id, user_id, balance, currency from wallets").
		WithArgs([]driver.Value{1}...).
		WillReturnRows(rows)

	for i := 0; i < b.N; i++ {
		repo.GetByID(ctx, 1)
	}
}

func BenchmarkTransfer(b *testing.B) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("cant create mock: %s", err)
	}
	defer sqlDB.Close()
	ctx := context.Background()

	walletOperation := operations.NewWalletOperationRepo(sqlDB)
	repo := NewWalletService(sqlDB, walletOperation)

	sourceOperationRows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	destinationOperationRows := sqlmock.NewRows([]string{"id"}).AddRow(2)

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
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.
		ExpectQuery("insert into wallet_operations").
		WithArgs([]driver.Value{operations.Deposit, 1, 2, decimal.NewFromInt(10)}...).
		WillReturnRows(sourceOperationRows)

	mock.
		ExpectQuery("insert into wallet_operations").
		WithArgs([]driver.Value{operations.Withdrawal, 2, 1, decimal.NewFromInt(10)}...).
		WillReturnRows(destinationOperationRows)

	// Commit transfer funds transaction
	mock.ExpectCommit()

	// Unlock operation
	mock.
		ExpectExec("select pg_advisory_unlock").
		WithArgs(TransferFunds).
		WillReturnResult(sqlmock.NewResult(2, 2))

	for i := 0; i < b.N; i++ {
		repo.Transfer(ctx, 1, 2, decimal.NewFromInt(10))
	}
}
