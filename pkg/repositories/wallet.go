package repositories

import (
	"billing_system_test_task/pkg/entities"
	"billing_system_test_task/pkg/operations"
	"context"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
)

const (
	CreateWallet = iota + 1
	EnrollWallet
	TransferFunds
)

// WalletsManager represents communication with wallets
type WalletsManager interface {
	Create(ctx context.Context, tx *sql.Tx, userID int64) (int64, error)
	Enroll(ctx context.Context, walletID int, amount decimal.Decimal) (int, error)
	GetByUserId(ctx context.Context, userID int) (*entities.Wallet, error)
	GetByID(ctx context.Context, walletID int) (*entities.Wallet, error)
	Transfer(ctx context.Context, walletFrom, walletTo int, amount decimal.Decimal) (int, error)
}

// WalletService shows structure for service of wallets
type WalletService struct {
	db *sql.DB
}

// NewWalletService returns instance of WalletService
func NewWalletService(db *sql.DB, walletOperationRepo operations.OperationsManager) WalletsManager {
	return WalletService{
		db: db,
	}
}

// Create creates new wallet for user
func (ws WalletService) Create(ctx context.Context, tx *sql.Tx, userID int64) (int64, error) {
	var (
		walletID int64
	)

	stmt, insertErr := tx.QueryContext(
		ctx,
		"insert into wallets(user_id) values($1) returning id",
		userID,
	)

	if insertErr != nil {
		return 0, fmt.Errorf("error wallet creation: %s", insertErr)
	}

	for stmt.Next() {
		scanErr := stmt.Scan(&walletID)
		if scanErr != nil {
			return 0, fmt.Errorf("error wallet id retrieving: %s", scanErr)
		}
	}

	// _, walletOperationErr := ws.walletOperationRepo.Create(ctx, tx, operations.Create, 0, int(walletID), walletBalance)
	// if walletOperationErr != nil {
	// 	log.Printf("Error of creating 'Create' wallet operation: %s", walletOperationErr.Error())
	// }

	return walletID, nil
}

// Enroll updates wallet's balance
func (ws WalletService) Enroll(ctx context.Context, walletID int, amount decimal.Decimal) (int, error) {
	// Check if amount is less or equal to 0
	if amount.LessThanOrEqual(decimal.Zero) {
		return 0, fmt.Errorf("amount should be greater than 0")
	}

	// Get connection from the pool
	// conn, _ := ws.db.Conn(ctx)

	// Apply AdvisoryLock for operation
	// _, alErr := conn.ExecContext(ctx, "select pg_advisory_lock($1)", EnrollWallet)
	// if alErr != nil {
	// 	return 0, fmt.Errorf("error of starting advisory lock: %s", alErr)
	// }

	// Begin transaction
	// tx, txErr := conn.BeginTx(ctx, nil)
	// if txErr != nil {
	// 	return 0, fmt.Errorf("error of transaction initialization: %s", txErr)
	// }
	// _, _ = tx.ExecContext(ctx, "set transaction isolation level serializable")

	// Update wallet 'balance' column
	_, updateErr := ws.db.ExecContext(ctx, "update wallets set balance=balance+$1 where id=$2 returning balance", amount, walletID)
	if updateErr != nil {
		// _ = tx.Rollback()
		// _, _ = conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, EnrollWallet)
		return 0, fmt.Errorf("error wallet enrollment: %s", updateErr)
	}

	// _, walletOperationErr := ws.walletOperationRepo.Create(ctx, tx, operations.Deposit, 0, int(walletID), amount)
	// if walletOperationErr != nil {
	// 	log.Printf("Error of creating 'Create' wallet operation: %s", walletOperationErr.Error())
	// }

	// txCommitErr := tx.Commit()
	// if txCommitErr != nil {
	// _ = tx.Rollback()
	// _, _ = conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, EnrollWallet)
	// return 0, fmt.Errorf("error of transaction commit: %s", txCommitErr)
	// }

	// _, auErr := conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, EnrollWallet)
	// if auErr != nil {
	// 	return 0, fmt.Errorf(
	// 		"error of unlocking user's %d postgres lock: %s",
	// 		EnrollWallet,
	// 		auErr,
	// 	)
	// }
	// conn.Close()

	return walletID, nil
}

// GetByID retrieves wallet by its ID
func (ws WalletService) GetByID(ctx context.Context, walletID int) (*entities.Wallet, error) {
	wallet := entities.Wallet{}
	getWalletErr := ws.db.
		QueryRowContext(ctx, "select id, user_id, balance, currency from wallets where id=$1", walletID).
		Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.Currency)
	if getWalletErr != nil {
		return nil, getWalletErr
	}
	return &wallet, nil
}

// GetByUserId retrieves wallet by user ID
func (ws WalletService) GetByUserId(ctx context.Context, userID int) (*entities.Wallet, error) {
	wallet := entities.Wallet{}
	getWalletErr := ws.db.
		QueryRowContext(ctx, "select id, user_id, balance, currency from wallets where user_id=$1", userID).
		Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.Currency)
	if getWalletErr != nil {
		return nil, getWalletErr
	}
	return &wallet, nil
}

// Transfer moves financial resources from one wallet to another
func (ws WalletService) Transfer(ctx context.Context, walletFrom, walletTo int, amount decimal.Decimal) (int, error) {
	// Receive source wallet
	sourceWallet, getSourceWalletErr := ws.GetByID(ctx, walletFrom)
	if getSourceWalletErr != nil {
		return 0, fmt.Errorf("error of receiving source wallet data: %s", getSourceWalletErr)
	}

	if sourceWallet.Balance.LessThanOrEqual(decimal.Zero) {
		return 0, fmt.Errorf("source wallet balance is less or equal to zero")
	}

	// Get connection from the pool
	// conn, _ := ws.db.Conn(ctx)

	// // Apply AdvisoryLock for operation
	// _, alErr := conn.ExecContext(ctx, "select pg_advisory_lock($1)", TransferFunds)
	// if alErr != nil {
	// 	return 0, fmt.Errorf("error of starting advisory lock: %s", alErr)
	// }

	// // Begin transaction
	// tx, txErr := conn.BeginTx(ctx, nil)
	// if txErr != nil {
	// 	return 0, fmt.Errorf("error of transaction initialization: %s", txErr)
	// }
	// _, _ = tx.ExecContext(ctx, "set transaction isolation level serializable")

	// Update source wallet 'balance' column
	_, updateSourceErr := ws.db.ExecContext(ctx, "update wallets set balance=balance-$1 where id=$2", amount, walletFrom)
	if updateSourceErr != nil {
		// _ = tx.Rollback()
		// _, _ = conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, TransferFunds)
		return 0, fmt.Errorf("error source wallet debit: %s", updateSourceErr)
	}

	// Update target wallet 'balance' column
	_, updateTargetErr := ws.db.ExecContext(ctx, "update wallets set balance=balance+$1 where id=$2", amount, walletTo)
	if updateTargetErr != nil {
		// _ = tx.Rollback()
		// _, _ = conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, TransferFunds)
		return 0, fmt.Errorf("error target wallet transfer: %s", updateTargetErr)
	}

	// _, walletDepositOperationErr := ws.walletOperationRepo.Create(ctx, tx, operations.Deposit, walletFrom, walletTo, amount)
	// if walletDepositOperationErr != nil {
	// 	log.Printf("Error of creating 'Deposit' wallet operation: %s", walletDepositOperationErr.Error())
	// }
	// _, walletWithdrawalOperationErr := ws.walletOperationRepo.Create(ctx, tx, operations.Withdrawal, walletTo, walletFrom, amount)
	// if walletWithdrawalOperationErr != nil {
	// 	log.Printf("Error of creating 'Withdrawal' wallet operation: %s", walletWithdrawalOperationErr.Error())
	// }

	// Commit transaction
	// txCommitErr := tx.Commit()
	// if txCommitErr != nil {
	// 	_ = tx.Rollback()
	// 	_, _ = conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, TransferFunds)
	// 	return 0, fmt.Errorf("error of transaction commit: %s", txCommitErr)
	// }

	// Perform advisory unlock
	// _, auErr := conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, TransferFunds)
	// if auErr != nil {
	// 	return 0, fmt.Errorf(
	// 		"error of unlocking user's %d postgres lock: %s",
	// 		TransferFunds,
	// 		auErr,
	// 	)
	// }
	// conn.Close()

	return walletFrom, nil
}