package usecases

import (
	"billing_system_test_task/pkg/adapters/tx"
	"billing_system_test_task/pkg/entities"
	"billing_system_test_task/pkg/repositories"
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
)

type userUsecaseTest struct {
	name                string
	args                []driver.Value
	funcName            string
	mockQuery           func(ctx context.Context, mockUserRepo *repositories.MockUsersManager, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx)
	err                 error
	expectedResultMatch func(actual interface{}) bool
}

var userUsecaseTests = []userUsecaseTest{
	userUsecaseTest{
		name:     "Success user creation",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(ctx context.Context, mockUserRepo *repositories.MockUsersManager, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			// Start users create transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			// Exec insert users query
			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			mockUserRepo.EXPECT().Create(ctx, "example@mail.com").Return(int64(1), nil)

			// Exec insert wallets query
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().Create(ctx, int64(1)).Return(int64(1), nil)

			mockOperationRepo.EXPECT().WithTx(txMock).Return(mockOperationRepo)
			mockOperationRepo.EXPECT().Create(ctx, repositories.Create, 0, 1, decimal.NewFromInt(0)).Return(1, nil)

			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			mockUserRepo.EXPECT().GetByWalletID(ctx, 1).Return(&entities.User{
				ID:    1,
				Email: "test@example.com",
				Wallet: &entities.Wallet{
					ID:       1,
					Balance:  decimal.NewFromInt(100),
					Currency: "USD",
				},
			}, nil)

			// Commit users create transaction
			txMock.EXPECT().Commit().Return(nil)
		},
		expectedResultMatch: func(actual interface{}) bool {
			user := entities.User{
				ID:    1,
				Email: "test@example.com",
				Wallet: &entities.Wallet{
					ID:       1,
					Balance:  decimal.NewFromInt(100),
					Currency: "USD",
				},
			}
			actualUser := actual.(*entities.User)
			return actualUser.ID == user.ID
		},
	},
	userUsecaseTest{
		name:     "Failed user creation (begin transaction error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(ctx context.Context, mockUserRepo *repositories.MockUsersManager, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			// Start users create transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(nil, fmt.Errorf("Transaction error"))
		},
		err: fmt.Errorf("Transaction error"),
	},
	userUsecaseTest{
		name:     "Failed user creation (user creation error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(ctx context.Context, mockUserRepo *repositories.MockUsersManager, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			// Start users create transaction
			// Start users create transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			// Exec insert users query
			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			mockUserRepo.EXPECT().Create(ctx, "example@mail.com").Return(int64(0), fmt.Errorf("create user error"))

			txMock.EXPECT().Rollback().Return(nil)
		},
		err: fmt.Errorf("create user error"),
	},
	userUsecaseTest{
		name:     "Failed user creation (wallet creation error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(ctx context.Context, mockUserRepo *repositories.MockUsersManager, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			// Start users create transaction
			// Start users create transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			// Exec insert users query
			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			mockUserRepo.EXPECT().Create(ctx, "example@mail.com").Return(int64(1), nil)

			// Exec insert wallets query
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().Create(ctx, int64(1)).Return(int64(0), fmt.Errorf("create wallet error"))

			txMock.EXPECT().Rollback().Return(nil)
		},
		err: fmt.Errorf("create wallet error"),
	},
	userUsecaseTest{
		name:     "Failed user creation (wallet operation creation error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(ctx context.Context, mockUserRepo *repositories.MockUsersManager, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			// Start users create transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			// Exec insert users query
			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			mockUserRepo.EXPECT().Create(ctx, "example@mail.com").Return(int64(1), nil)

			// Exec insert wallets query
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().Create(ctx, int64(1)).Return(int64(1), nil)

			mockOperationRepo.EXPECT().WithTx(txMock).Return(mockOperationRepo)
			mockOperationRepo.EXPECT().Create(ctx, repositories.Create, 0, 1, decimal.NewFromInt(0)).Return(0, fmt.Errorf("create wallet operation error"))

			txMock.EXPECT().Rollback().Return(nil)
		},
		err: fmt.Errorf("create wallet operation error"),
	},
	userUsecaseTest{
		name:     "Failed user creation (get user error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(ctx context.Context, mockUserRepo *repositories.MockUsersManager, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			// Start users create transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			// Exec insert users query
			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			mockUserRepo.EXPECT().Create(ctx, "example@mail.com").Return(int64(1), nil)

			// Exec insert wallets query
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().Create(ctx, int64(1)).Return(int64(1), nil)

			mockOperationRepo.EXPECT().WithTx(txMock).Return(mockOperationRepo)
			mockOperationRepo.EXPECT().Create(ctx, repositories.Create, 0, 1, decimal.NewFromInt(0)).Return(1, nil)

			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			mockUserRepo.EXPECT().GetByWalletID(ctx, 1).Return(nil, fmt.Errorf("get user error"))

			txMock.EXPECT().Rollback().Return(nil)
		},
		err: fmt.Errorf("get user error"),
	},
	userUsecaseTest{
		name:     "Failed user creation (transaction commit error)",
		funcName: "Create",
		args:     []driver.Value{"example@mail.com"},
		mockQuery: func(ctx context.Context, mockUserRepo *repositories.MockUsersManager, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			// Start users create transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			// Exec insert users query
			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			mockUserRepo.EXPECT().Create(ctx, "example@mail.com").Return(int64(1), nil)

			// Exec insert wallets query
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().Create(ctx, int64(1)).Return(int64(1), nil)

			mockOperationRepo.EXPECT().WithTx(txMock).Return(mockOperationRepo)
			mockOperationRepo.EXPECT().Create(ctx, repositories.Create, 0, 1, decimal.NewFromInt(0)).Return(1, nil)

			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			mockUserRepo.EXPECT().GetByWalletID(ctx, 1).Return(&entities.User{
				ID:    1,
				Email: "test@example.com",
				Wallet: &entities.Wallet{
					ID:       1,
					Balance:  decimal.NewFromInt(100),
					Currency: "USD",
				},
			}, nil)

			// Commit users create transaction
			txMock.EXPECT().Commit().Return(fmt.Errorf("transaction commit error"))
		},
		err: fmt.Errorf("transaction commit error"),
	},
}

func TestUserUsecase(t *testing.T) {
	for _, tc := range userUsecaseTests {
		ctrl := gomock.NewController(t)
		ctx := context.Background()
		realArgs := []reflect.Value{
			reflect.ValueOf(ctx),
		}
		db, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("cant create mock: %s", err)
		}
		defer db.Close()

		txManager := tx.NewMockTxBeginner(ctrl)
		txMock := tx.NewMockTx(ctrl)
		walletsRepo := repositories.NewMockWalletsManager(ctrl)
		usersRepo := repositories.NewMockUsersManager(ctrl)
		operationsRepo := repositories.NewMockOperationsManager(ctrl)

		interactor := NewUserInteractor(usersRepo, walletsRepo, operationsRepo, txManager).(UserInteractor)

		for _, arg := range tc.args {
			realArgs = append(realArgs, reflect.ValueOf(arg))
		}
		tc.mockQuery(ctx, usersRepo, walletsRepo, operationsRepo, txManager, txMock)

		var result []reflect.Value
		if len(tc.args) > 0 {
			result = reflect.ValueOf(
				&interactor,
			).MethodByName(
				tc.funcName,
			).Call(realArgs)
		} else {
			result = reflect.ValueOf(
				&interactor,
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
			resultErr := rerr.(error)
			if tc.err.Error() != resultErr.Error() {
				t.Errorf("errors do not match. Expected '%s', got '%s'", tc.err, rerr)
				return
			}
		}

		if tc.err == nil && !tc.expectedResultMatch(resultValue) {
			t.Errorf("result data is not matched. Got %s", resultValue)
		}
	}
}
