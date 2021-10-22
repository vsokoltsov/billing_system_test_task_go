package usecases

import (
	"billing_system_test_task/internal/adapters"
	"billing_system_test_task/internal/adapters/tx"
	"billing_system_test_task/internal/entities"
	"billing_system_test_task/internal/repositories"
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
			txMock.EXPECT().Rollback().Return(nil)
		},
		err: fmt.Errorf("transaction commit error"),
	},
	userUsecaseTest{
		name:     "Success user's wallet enrollment",
		funcName: "Enroll",
		args:     []driver.Value{1, decimal.NewFromInt(10)},
		mockQuery: func(ctx context.Context, mockUserRepo *repositories.MockUsersManager, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			user := &entities.User{
				ID:    1,
				Email: "test@example.com",
				Wallet: &entities.Wallet{
					ID:       1,
					Balance:  decimal.NewFromInt(100),
					Currency: "USD",
				},
			}
			// Start users enroll transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)
			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			mockUserRepo.EXPECT().GetByID(ctx, 1).Return(user, nil)

			// Exec insert wallets query
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().Enroll(ctx, 1, decimal.NewFromInt(10)).Return(1, nil)

			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			user.Wallet.Balance = user.Wallet.Balance.Add(decimal.NewFromInt(10))
			mockUserRepo.EXPECT().GetByWalletID(ctx, 1).Return(user, nil)

			// Commit users create transaction
			txMock.EXPECT().Commit().Return(nil)
		},
		expectedResultMatch: func(actual interface{}) bool {
			user := entities.User{
				ID:    1,
				Email: "test@example.com",
				Wallet: &entities.Wallet{
					ID:       1,
					Balance:  decimal.NewFromInt(110),
					Currency: "USD",
				},
			}
			actualUser := actual.(*entities.User)
			return actualUser.ID == user.ID
		},
	},
	userUsecaseTest{
		name:     "Failed user's wallet enrollment (transaction begin)",
		funcName: "Enroll",
		args:     []driver.Value{1, decimal.NewFromInt(10)},
		mockQuery: func(ctx context.Context, mockUserRepo *repositories.MockUsersManager, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			// Start users enroll transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(nil, fmt.Errorf("tx start error"))
		},
		err: fmt.Errorf("tx start error"),
	},
	userUsecaseTest{
		name:     "Failed user's wallet enrollment (GetByID error)",
		funcName: "Enroll",
		args:     []driver.Value{1, decimal.NewFromInt(10)},
		mockQuery: func(ctx context.Context, mockUserRepo *repositories.MockUsersManager, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			// Start users enroll transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			mockUserRepo.EXPECT().GetByID(ctx, 1).Return(nil, fmt.Errorf("tx start error"))
			txMock.EXPECT().Rollback().Return(nil)
		},
		err: fmt.Errorf("tx start error"),
	},
	userUsecaseTest{
		name:     "Failed user's wallet enrollment",
		funcName: "Enroll",
		args:     []driver.Value{1, decimal.NewFromInt(10)},
		mockQuery: func(ctx context.Context, mockUserRepo *repositories.MockUsersManager, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			user := &entities.User{
				ID:    1,
				Email: "test@example.com",
				Wallet: &entities.Wallet{
					ID:       1,
					Balance:  decimal.NewFromInt(100),
					Currency: "USD",
				},
			}
			// Start users enroll transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)
			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			mockUserRepo.EXPECT().GetByID(ctx, 1).Return(user, nil)

			// Exec insert wallets query
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().Enroll(ctx, 1, decimal.NewFromInt(10)).Return(0, fmt.Errorf("Enroll error"))
			txMock.EXPECT().Rollback().Return(nil)
		},
		err: fmt.Errorf("Enroll error"),
	},
	userUsecaseTest{
		name:     "Failed user's wallet enrollment (GetByWalletID error)",
		funcName: "Enroll",
		args:     []driver.Value{1, decimal.NewFromInt(10)},
		mockQuery: func(ctx context.Context, mockUserRepo *repositories.MockUsersManager, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			user := &entities.User{
				ID:    1,
				Email: "test@example.com",
				Wallet: &entities.Wallet{
					ID:       1,
					Balance:  decimal.NewFromInt(100),
					Currency: "USD",
				},
			}
			// Start users enroll transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)
			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			mockUserRepo.EXPECT().GetByID(ctx, 1).Return(user, nil)

			// Exec insert wallets query
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().Enroll(ctx, 1, decimal.NewFromInt(10)).Return(1, nil)

			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			user.Wallet.Balance = user.Wallet.Balance.Add(decimal.NewFromInt(10))
			mockUserRepo.EXPECT().GetByWalletID(ctx, 1).Return(nil, fmt.Errorf("GetByWalletID error"))

			// Commit users create transaction
			txMock.EXPECT().Rollback().Return(nil)
		},
		err: fmt.Errorf("GetByWalletID error"),
	},
	userUsecaseTest{
		name:     "Failed user's wallet enrollment (Commit error)",
		funcName: "Enroll",
		args:     []driver.Value{1, decimal.NewFromInt(10)},
		mockQuery: func(ctx context.Context, mockUserRepo *repositories.MockUsersManager, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			user := &entities.User{
				ID:    1,
				Email: "test@example.com",
				Wallet: &entities.Wallet{
					ID:       1,
					Balance:  decimal.NewFromInt(100),
					Currency: "USD",
				},
			}
			// Start users enroll transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)
			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			mockUserRepo.EXPECT().GetByID(ctx, 1).Return(user, nil)

			// Exec insert wallets query
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().Enroll(ctx, 1, decimal.NewFromInt(10)).Return(1, nil)

			mockUserRepo.EXPECT().WithTx(txMock).Return(mockUserRepo)
			user.Wallet.Balance = user.Wallet.Balance.Add(decimal.NewFromInt(10))
			mockUserRepo.EXPECT().GetByWalletID(ctx, 1).Return(user, nil)

			// Commit users create transaction
			txMock.EXPECT().Commit().Return(fmt.Errorf("commit error"))
			txMock.EXPECT().Rollback().Return(nil)
		},
		err: fmt.Errorf("commit error"),
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

		errFactory := adapters.NewHTTPErrorsFactory()
		txManager := tx.NewMockTxBeginner(ctrl)
		txMock := tx.NewMockTx(ctrl)
		walletsRepo := repositories.NewMockWalletsManager(ctrl)
		usersRepo := repositories.NewMockUsersManager(ctrl)
		operationsRepo := repositories.NewMockOperationsManager(ctrl)

		interactor := NewUserInteractor(usersRepo, walletsRepo, operationsRepo, txManager, errFactory)

		for _, arg := range tc.args {
			realArgs = append(realArgs, reflect.ValueOf(arg))
		}
		tc.mockQuery(ctx, usersRepo, walletsRepo, operationsRepo, txManager, txMock)

		var result []reflect.Value
		if len(tc.args) > 0 {
			result = reflect.ValueOf(
				interactor,
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
			reflectErr *adapters.HTTPError
		)
		resultValue := result[0].Interface()
		rerr := result[1].Interface()
		if rerr != nil {
			reflectErr = rerr.(*adapters.HTTPError)
		}

		if reflectErr != nil && tc.err == nil {
			t.Errorf("unexpected err: %s", reflectErr.GetError().Error())
			return
		}

		if tc.err != nil {
			if reflectErr == nil {
				t.Errorf("expected error, got nil: %s", reflectErr.GetError().Error())
				return
			}
			resultErr := rerr.(*adapters.HTTPError)
			if tc.err.Error() != resultErr.GetError().Error() {
				t.Errorf("errors do not match. Expected '%s', got '%s'", tc.err, rerr)
				return
			}
		}

		if tc.err == nil && !tc.expectedResultMatch(resultValue) {
			t.Errorf("result data is not matched. Got %s", resultValue)
		}
	}
}
