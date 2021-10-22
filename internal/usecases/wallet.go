package usecases

import (
	"billing_system_test_task/internal/adapters"
	"billing_system_test_task/internal/adapters/tx"
	"billing_system_test_task/internal/repositories"
	"context"

	"github.com/shopspring/decimal"
)

// WalletUseCase represents contracts for wallet's use cases
type WalletUseCase interface {
	Transfer(ctx context.Context, walletFrom, walletTo int, amount decimal.Decimal) (int, adapters.Error)
}

type WalletInteractor struct {
	walletRepo        repositories.WalletsManager
	errFactory        adapters.ErrorsFactory
	txManager         tx.TxBeginner
	operationsManager repositories.OperationsManager
}

func NewWalletInteractor(walletRepo repositories.WalletsManager, operationsManager repositories.OperationsManager, errFactory adapters.ErrorsFactory, txManager tx.TxBeginner) WalletUseCase {
	return &WalletInteractor{
		walletRepo:        walletRepo,
		errFactory:        errFactory,
		txManager:         txManager,
		operationsManager: operationsManager,
	}
}

func (wi *WalletInteractor) Transfer(ctx context.Context, walletFrom, walletTo int, amount decimal.Decimal) (int, adapters.Error) {
	// Start transaction
	tx, txErr := wi.txManager.BeginTrx(ctx, nil)
	if txErr != nil {
		return 0, wi.errFactory.DefaultError(txErr)
	}

	// Perform transfer
	walletSourceID, transferErr := wi.walletRepo.WithTx(tx).Transfer(
		ctx,
		walletFrom,
		walletTo,
		amount,
	)
	if transferErr != nil {
		return 0, wi.errFactory.DefaultError(transferErr)
	}

	txWalletOpRepo := wi.operationsManager.WithTx(tx)
	// Create wallet operation instance for deposit
	_, depositOpErrr := txWalletOpRepo.Create(ctx, repositories.Deposit, walletFrom, walletTo, amount)
	if depositOpErrr != nil {
		return 0, wi.errFactory.DefaultError(depositOpErrr)
	}

	// Create wallet operation instance for withdrawal
	_, withdrawalOpErrr := txWalletOpRepo.Create(ctx, repositories.Withdrawal, walletTo, walletFrom, amount)
	if withdrawalOpErrr != nil {
		return 0, wi.errFactory.DefaultError(withdrawalOpErrr)
	}

	// Commit transaction
	if commitErr := tx.Commit(); commitErr != nil {
		_ = tx.Rollback()
		return 0, wi.errFactory.DefaultError(commitErr)
	}
	return walletSourceID, nil
}
