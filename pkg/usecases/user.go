package usecases

import (
	"billing_system_test_task/pkg/adapters/tx"
	"billing_system_test_task/pkg/entities"
	"billing_system_test_task/pkg/repositories"
	"context"

	"github.com/shopspring/decimal"
)

type UserUseCase interface {
	Create(ctx context.Context, email string) (*entities.User, error)
}

type UserInteractor struct {
	userRepo          repositories.UsersManager
	walletsRepo       repositories.WalletsManager
	operationsManager repositories.OperationsManager
	txManager         tx.TxBeginner
}

func NewUserInteractor(userRepo repositories.UsersManager, walletsRepo repositories.WalletsManager, operationsManager repositories.OperationsManager, txManager tx.TxBeginner) UserUseCase {
	return UserInteractor{
		userRepo:          userRepo,
		walletsRepo:       walletsRepo,
		txManager:         txManager,
		operationsManager: operationsManager,
	}
}

// Create creates new user, its wallet and operation for that event
func (ui UserInteractor) Create(ctx context.Context, email string) (*entities.User, error) {

	tx, txErr := ui.txManager.BeginTrx(ctx, nil)
	if txErr != nil {
		return nil, txErr
	}

	txUserRepo := ui.userRepo.WithTx(tx)
	userID, userErr := txUserRepo.Create(ctx, email)
	if userErr != nil {
		_ = tx.Rollback()
		return nil, userErr
	}

	walletID, walletErr := ui.walletsRepo.WithTx(tx).Create(ctx, userID)
	if walletErr != nil {
		_ = tx.Rollback()
		return nil, walletErr
	}

	_, walletOperationErr := ui.operationsManager.WithTx(tx).Create(ctx, repositories.Create, 0, int(walletID), decimal.NewFromInt(0))
	if walletOperationErr != nil {
		_ = tx.Rollback()
		return nil, walletOperationErr
	}

	user, getUserErr := txUserRepo.GetByWalletID(ctx, int(walletID))
	if getUserErr != nil {
		return nil, getUserErr
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return nil, commitErr
	}

	return user, nil
}
