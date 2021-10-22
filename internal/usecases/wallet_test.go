package usecases

import (
	"billing_system_test_task/internal/adapters"
	"billing_system_test_task/internal/adapters/tx"
	"billing_system_test_task/internal/entities"
	"billing_system_test_task/internal/repositories"
	"context"
	"database/sql/driver"
	"fmt"
	reflect "reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	gomock "github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
)

type walletUsecaseTest struct {
	name                string
	args                []driver.Value
	funcName            string
	mockQuery           func(ctx context.Context, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx)
	err                 error
	expectedResultMatch func(actual interface{}) bool
}

var walletUseCases = []walletUsecaseTest{
	walletUsecaseTest{
		name:     "Success wallet transfer",
		args:     []driver.Value{1, 2, decimal.NewFromInt(10)},
		funcName: "Transfer",
		mockQuery: func(ctx context.Context, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			sourceWallet := &entities.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  decimal.NewFromInt(100),
				Currency: "USD",
			}
			destinationWallet := &entities.Wallet{
				ID:       2,
				UserID:   2,
				Balance:  decimal.NewFromInt(100),
				Currency: "USD",
			}
			// Start wallet transfer transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			// Receive source wallet
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().GetByID(ctx, sourceWallet.ID).Return(sourceWallet, nil)

			// Receive destination wallet
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().GetByID(ctx, destinationWallet.ID).Return(destinationWallet, nil)

			// Perform transfer itself
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().Transfer(ctx, 1, 2, decimal.NewFromInt(10)).Return(1, nil)

			// Create operation for the wallet's deposit
			mockOperationRepo.EXPECT().WithTx(txMock).Return(mockOperationRepo)
			mockOperationRepo.EXPECT().Create(ctx, repositories.Deposit, 1, 2, decimal.NewFromInt(10)).Return(1, nil)

			// Create operation for the wallet's withdrawal
			mockOperationRepo.EXPECT().WithTx(txMock).Return(mockOperationRepo)
			mockOperationRepo.EXPECT().Create(ctx, repositories.Withdrawal, 2, 1, decimal.NewFromInt(10)).Return(2, nil)

			// Commit wallet transfer transaction
			txMock.EXPECT().Commit().Return(nil)

		},
		expectedResultMatch: func(actual interface{}) bool {
			return true
		},
	},
	walletUsecaseTest{
		name:     "Failed wallet transfer (get source wallet error)",
		args:     []driver.Value{1, 2, decimal.NewFromInt(10)},
		funcName: "Transfer",
		mockQuery: func(ctx context.Context, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			sourceWallet := &entities.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  decimal.NewFromInt(100),
				Currency: "USD",
			}
			// Start wallet transfer transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			// Receive source wallet
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().GetByID(ctx, sourceWallet.ID).Return(nil, fmt.Errorf("source wallet error"))

			// Rollback wallet transfer transaction
			txMock.EXPECT().Rollback().Return(nil)

		},
		err: fmt.Errorf("source wallet error"),
	},
	walletUsecaseTest{
		name:     "Failed wallet transfer (source wallet balance is 0)",
		args:     []driver.Value{1, 2, decimal.NewFromInt(10)},
		funcName: "Transfer",
		mockQuery: func(ctx context.Context, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			sourceWallet := &entities.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  decimal.NewFromInt(00),
				Currency: "USD",
			}
			// Start wallet transfer transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			// Receive source wallet
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().GetByID(ctx, sourceWallet.ID).Return(sourceWallet, nil)

			// Rollback wallet transfer transaction
			txMock.EXPECT().Rollback().Return(nil)

		},
		err: fmt.Errorf("source wallet balance is less or equal to zero"),
	},
	walletUsecaseTest{
		name:     "Failed wallet transfer (get destination wallet error)",
		args:     []driver.Value{1, 2, decimal.NewFromInt(10)},
		funcName: "Transfer",
		mockQuery: func(ctx context.Context, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			sourceWallet := &entities.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  decimal.NewFromInt(100),
				Currency: "USD",
			}
			destinationWallet := &entities.Wallet{
				ID:       2,
				UserID:   2,
				Balance:  decimal.NewFromInt(100),
				Currency: "USD",
			}
			// Start wallet transfer transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			// Receive source wallet
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().GetByID(ctx, sourceWallet.ID).Return(sourceWallet, nil)

			// Receive destination wallet
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().GetByID(ctx, destinationWallet.ID).Return(nil, fmt.Errorf("destination wallet error"))

			// Commit wallet transfer transaction
			txMock.EXPECT().Rollback().Return(nil)

		},
		err: fmt.Errorf("destination wallet error"),
	},
	walletUsecaseTest{
		name:     "Failed wallet transfer (start transaction error)",
		args:     []driver.Value{1, 2, decimal.NewFromInt(10)},
		funcName: "Transfer",
		mockQuery: func(ctx context.Context, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			// Start wallet transfer transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(nil, fmt.Errorf("tx start error"))

		},
		err: fmt.Errorf("tx start error"),
	},
	walletUsecaseTest{
		name:     "Failed wallet transfer (Transfer error)",
		args:     []driver.Value{1, 2, decimal.NewFromInt(10)},
		funcName: "Transfer",
		mockQuery: func(ctx context.Context, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			sourceWallet := &entities.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  decimal.NewFromInt(100),
				Currency: "USD",
			}
			destinationWallet := &entities.Wallet{
				ID:       2,
				UserID:   2,
				Balance:  decimal.NewFromInt(100),
				Currency: "USD",
			}

			// Start wallet transfer transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			// Receive source wallet
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().GetByID(ctx, sourceWallet.ID).Return(sourceWallet, nil)

			// Receive destination wallet
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().GetByID(ctx, destinationWallet.ID).Return(destinationWallet, nil)

			// Perform transfer itself
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().Transfer(ctx, sourceWallet.ID, destinationWallet.ID, decimal.NewFromInt(10)).Return(0, fmt.Errorf("transfer error"))

			// Rollback wallet transfer transaction
			txMock.EXPECT().Rollback().Return(nil)

		},
		err: fmt.Errorf("transfer error"),
	},
	walletUsecaseTest{
		name:     "Failed wallet transfer (deposit operation error)",
		args:     []driver.Value{1, 2, decimal.NewFromInt(10)},
		funcName: "Transfer",
		mockQuery: func(ctx context.Context, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			sourceWallet := &entities.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  decimal.NewFromInt(100),
				Currency: "USD",
			}
			destinationWallet := &entities.Wallet{
				ID:       2,
				UserID:   2,
				Balance:  decimal.NewFromInt(100),
				Currency: "USD",
			}
			// Start wallet transfer transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			// Receive source wallet
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().GetByID(ctx, sourceWallet.ID).Return(sourceWallet, nil)

			// Receive destination wallet
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().GetByID(ctx, destinationWallet.ID).Return(destinationWallet, nil)

			// Perform transfer itself
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().Transfer(ctx, 1, 2, decimal.NewFromInt(10)).Return(1, nil)

			// Create operation for the wallet's deposit
			mockOperationRepo.EXPECT().WithTx(txMock).Return(mockOperationRepo)
			mockOperationRepo.EXPECT().Create(ctx, repositories.Deposit, 1, 2, decimal.NewFromInt(10)).Return(0, fmt.Errorf("depoit error"))

			// Rollback wallet transfer transaction
			txMock.EXPECT().Rollback().Return(nil)

		},
		err: fmt.Errorf("depoit error"),
	},
	walletUsecaseTest{
		name:     "Failed wallet transfer (withdrawal operation create error)",
		args:     []driver.Value{1, 2, decimal.NewFromInt(10)},
		funcName: "Transfer",
		mockQuery: func(ctx context.Context, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			sourceWallet := &entities.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  decimal.NewFromInt(100),
				Currency: "USD",
			}
			destinationWallet := &entities.Wallet{
				ID:       2,
				UserID:   2,
				Balance:  decimal.NewFromInt(100),
				Currency: "USD",
			}

			// Start wallet transfer transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			// Receive source wallet
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().GetByID(ctx, sourceWallet.ID).Return(sourceWallet, nil)

			// Receive destination wallet
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().GetByID(ctx, destinationWallet.ID).Return(destinationWallet, nil)

			// Perform transfer itself
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().Transfer(ctx, 1, 2, decimal.NewFromInt(10)).Return(1, nil)

			// Create operation for the wallet's deposit
			mockOperationRepo.EXPECT().WithTx(txMock).Return(mockOperationRepo)
			mockOperationRepo.EXPECT().Create(ctx, repositories.Deposit, 1, 2, decimal.NewFromInt(10)).Return(1, nil)

			// Create operation for the wallet's withdrawal
			mockOperationRepo.EXPECT().WithTx(txMock).Return(mockOperationRepo)
			mockOperationRepo.EXPECT().Create(ctx, repositories.Withdrawal, 2, 1, decimal.NewFromInt(10)).Return(0, fmt.Errorf("witdhdrawal error"))

			// Commit wallet transfer transaction
			txMock.EXPECT().Rollback().Return(nil)

		},
		err: fmt.Errorf("witdhdrawal error"),
	},
	walletUsecaseTest{
		name:     "Failed wallet transfer (tx commit error)",
		args:     []driver.Value{1, 2, decimal.NewFromInt(10)},
		funcName: "Transfer",
		mockQuery: func(ctx context.Context, mockWalletRepo *repositories.MockWalletsManager, mockOperationRepo *repositories.MockOperationsManager, mockTxManager *tx.MockTxBeginner, txMock *tx.MockTx) {
			sourceWallet := &entities.Wallet{
				ID:       1,
				UserID:   1,
				Balance:  decimal.NewFromInt(100),
				Currency: "USD",
			}
			destinationWallet := &entities.Wallet{
				ID:       2,
				UserID:   2,
				Balance:  decimal.NewFromInt(100),
				Currency: "USD",
			}

			// Start wallet transfer transaction
			mockTxManager.EXPECT().BeginTrx(ctx, nil).Return(txMock, nil)

			// Receive source wallet
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().GetByID(ctx, sourceWallet.ID).Return(sourceWallet, nil)

			// Receive destination wallet
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().GetByID(ctx, destinationWallet.ID).Return(destinationWallet, nil)

			// Perform transfer itself
			mockWalletRepo.EXPECT().WithTx(txMock).Return(mockWalletRepo)
			mockWalletRepo.EXPECT().Transfer(ctx, 1, 2, decimal.NewFromInt(10)).Return(1, nil)

			// Create operation for the wallet's deposit
			mockOperationRepo.EXPECT().WithTx(txMock).Return(mockOperationRepo)
			mockOperationRepo.EXPECT().Create(ctx, repositories.Deposit, 1, 2, decimal.NewFromInt(10)).Return(1, nil)

			// Create operation for the wallet's withdrawal
			mockOperationRepo.EXPECT().WithTx(txMock).Return(mockOperationRepo)
			mockOperationRepo.EXPECT().Create(ctx, repositories.Withdrawal, 2, 1, decimal.NewFromInt(10)).Return(2, nil)

			// Commit wallet transfer transaction
			txMock.EXPECT().Commit().Return(fmt.Errorf("tx commit err"))
			txMock.EXPECT().Rollback().Return(nil)

		},
		err: fmt.Errorf("tx commit err"),
	},
}

// Test usecases for wallet
func TestWalletUsecase(t *testing.T) {
	for _, tc := range walletUseCases {
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
		operationsRepo := repositories.NewMockOperationsManager(ctrl)

		interactor := NewWalletInteractor(walletsRepo, operationsRepo, errFactory, txManager)

		for _, arg := range tc.args {
			realArgs = append(realArgs, reflect.ValueOf(arg))
		}
		tc.mockQuery(ctx, walletsRepo, operationsRepo, txManager, txMock)

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
