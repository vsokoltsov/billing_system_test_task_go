package wallets

import (
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

// IWalletRepo represents communication with wallets
type IWalletRepo interface {
	Create(ctx context.Context, tx *sql.Tx, userID int64) (int64, error)
	Enroll(ctx context.Context, walletID int, amount decimal.Decimal) (int, error)
	GetByUserId(ctx context.Context, userID int) (*Wallet, error)
	GetByID(ctx context.Context, walletID int) (*Wallet, error)
	Transfer(ctx context.Context, walletFrom, walletTo int, amount decimal.Decimal) (int, error)
}

// WalletService shows structure for service of wallets
type WalletService struct {
	db                  *sql.DB
	walletOperationRepo operations.IWalletOperationRepo
}

// NewWalletService returns instance of WalletService
func NewWalletService(db *sql.DB, walletOperationRepo operations.IWalletOperationRepo) IWalletRepo {
	return WalletService{
		db:                  db,
		walletOperationRepo: walletOperationRepo,
	}
}

// Create creates new wallet for user
func (ws WalletService) Create(ctx context.Context, tx *sql.Tx, userID int64) (int64, error) {
	var walletID int64
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
			return 0, fmt.Errorf("error wallet id retrieving: %s", insertErr)
		}
	}

	return walletID, nil
}

// Enroll updates wallet's balance
func (ws WalletService) Enroll(ctx context.Context, walletID int, amount decimal.Decimal) (int, error) {

	// Check if amount is less or equal to 0
	if amount.LessThanOrEqual(decimal.Zero) {
		return 0, fmt.Errorf("amount should be greater than 0")
	}

	// Get connection from the pool
	conn, _ := ws.db.Conn(ctx)

	// Apply AdvisoryLock for operation
	_, alErr := conn.ExecContext(ctx, "select pg_advisory_lock($1)", EnrollWallet)
	if alErr != nil {
		return 0, fmt.Errorf("error of starting advisory lock: %s", alErr)
	}

	// Begin transaction
	tx, txErr := conn.BeginTx(ctx, nil)
	if txErr != nil {
		return 0, fmt.Errorf("error of transaction initialization: %s", txErr)
	}
	tx.ExecContext(ctx, "set transaction isolation level serializable")

	// Update wallet 'balance' column
	_, updateErr := tx.ExecContext(ctx, "update wallets set balance=balance+$1 where id=$2", amount, walletID)
	if updateErr != nil {
		tx.Rollback()
		conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, EnrollWallet)
		return 0, fmt.Errorf("error wallet enrollment: %s", updateErr)
	}

	txCommitErr := tx.Commit()
	if txCommitErr != nil {
		tx.Rollback()
		conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, EnrollWallet)
		return 0, fmt.Errorf("error of transaction commit: %s", txCommitErr)
	}

	_, auErr := conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, EnrollWallet)
	if auErr != nil {
		return 0, fmt.Errorf(
			"error of unlocking user's %d postgres lock: %s",
			EnrollWallet,
			auErr,
		)
	}
	conn.Close()

	return walletID, nil
}

// GetByID retrieves wallet by its ID
func (ws WalletService) GetByID(ctx context.Context, walletID int) (*Wallet, error) {
	wallet := Wallet{}
	getWalletErr := ws.db.
		QueryRowContext(ctx, "select id, user_id, balance, currency from wallets where id=$1", walletID).
		Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.Currency)
	if getWalletErr != nil {
		return nil, getWalletErr
	}
	return &wallet, nil
}

// GetByUserId retrieves wallet by user ID
func (ws WalletService) GetByUserId(ctx context.Context, userID int) (*Wallet, error) {
	wallet := Wallet{}
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
	conn, _ := ws.db.Conn(ctx)

	// Apply AdvisoryLock for operation
	_, alErr := conn.ExecContext(ctx, "select pg_advisory_lock($1)", TransferFunds)
	if alErr != nil {
		return 0, fmt.Errorf("error of starting advisory lock: %s", alErr)
	}

	// Begin transaction
	tx, txErr := conn.BeginTx(ctx, nil)
	if txErr != nil {
		return 0, fmt.Errorf("error of transaction initialization: %s", txErr)
	}
	tx.ExecContext(ctx, "set transaction isolation level serializable")

	// Update source wallet 'balance' column
	_, updateSourceErr := tx.ExecContext(ctx, "update wallets set balance=balance-$1 where id=$2", amount, walletFrom)
	if updateSourceErr != nil {
		tx.Rollback()
		conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, TransferFunds)
		return 0, fmt.Errorf("error source wallet debit: %s", updateSourceErr)
	}

	// Update target wallet 'balance' column
	_, updateTargetErr := tx.ExecContext(ctx, "update wallets set balance=balance+$1 where id=$2", amount, walletTo)
	if updateTargetErr != nil {
		tx.Rollback()
		conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, TransferFunds)
		return 0, fmt.Errorf("error target wallet transfer: %s", updateTargetErr)
	}

	// Commit transaction
	txCommitErr := tx.Commit()
	if txCommitErr != nil {
		tx.Rollback()
		conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, TransferFunds)
		return 0, fmt.Errorf("error of transaction commit: %s", txCommitErr)
	}

	// Perform advisory unlock
	_, auErr := conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, TransferFunds)
	if auErr != nil {
		return 0, fmt.Errorf(
			"error of unlocking user's %d postgres lock: %s",
			TransferFunds,
			auErr,
		)
	}
	conn.Close()

	return walletFrom, nil
}
