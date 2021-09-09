package operations

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

type operationRepoTestCase struct {
	name      string
	funcName  string
	mockQuery func(mock sqlmock.Sqlmock)
	args      []driver.Value
	err       error
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
			rows := sqlmock.NewRows([]string{"id"})
			rows = rows.AddRow(1)

			// Exec insert wallets
			mock.
				ExpectQuery("select id, operation, wallet_from, wallet_to, amount, created_at from wallet_operations").
				WillReturnRows(rows)

		},
	},
	operationRepoTestCase{
		name:     "Success receiving list of items with paging",
		funcName: "List",
		args:     []driver.Value{&ListParams{page: 1, perPage: 10}},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id"})
			rows = rows.AddRow(1)

			// Exec insert wallets
			mock.
				ExpectQuery("select id, operation, wallet_from, wallet_to, amount, created_at from wallet_operations").
				WithArgs([]driver.Value{0, 10}...).
				WillReturnRows(rows)

		},
	},
	operationRepoTestCase{
		name:     "Success receiving list of items with paging more than 1",
		funcName: "List",
		args:     []driver.Value{&ListParams{page: 3, perPage: 10}},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id"})
			rows = rows.AddRow(1)

			// Exec insert wallets
			mock.
				ExpectQuery("select id, operation, wallet_from, wallet_to, amount, created_at from wallet_operations").
				WithArgs([]driver.Value{20, 10}...).
				WillReturnRows(rows)

		},
	},
	operationRepoTestCase{
		name:     "Success receiving list of items date filtering",
		funcName: "List",
		args:     []driver.Value{&ListParams{date: "2020-01-01"}},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id"})
			rows = rows.AddRow(1)

			// Exec insert wallets
			mock.
				ExpectQuery("where created_at = to_date").
				WithArgs([]driver.Value{"2020-01-01"}...).
				WillReturnRows(rows)

		},
	},
	operationRepoTestCase{
		name:     "Success receiving list of items with all parameters",
		funcName: "List",
		args:     []driver.Value{&ListParams{page: 1, perPage: 10, date: "2020-01-01"}},
		mockQuery: func(mock sqlmock.Sqlmock) {
			// Begin transaction
			rows := sqlmock.NewRows([]string{"id"})
			rows = rows.AddRow(1)

			// Exec insert wallets
			mock.
				ExpectQuery("where created_at = to_date").
				WithArgs([]driver.Value{"2020-01-01", 0, 10}...).
				WillReturnRows(rows)

		},
	},
}

func TestWalletRepo(t *testing.T) {
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
	walletOperation := NewWalletOperationRepo(db)
	_, correctType := walletOperation.(WalletOperationService)
	if !correctType {
		t.Errorf("Wrong type of WalletOperationService")
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

	rows := sqlmock.NewRows([]string{"id"})
	rows = rows.AddRow(1)

	mock.
		ExpectQuery("insert into wallet_operations").
		WithArgs([]driver.Value{Create, 1, 2, decimal.NewFromInt(0)}...).
		WillReturnRows(rows)

	repo := NewWalletOperationRepo(sqlDB)
	for i := 0; i < b.N; i++ {
		repo.Create(ctx, tx, Create, 1, 2, decimal.NewFromInt(0))
	}
}

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
		repo.List(ctx, nil)
	}
}
