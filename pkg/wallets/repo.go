package wallets

import (
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

type SQLRepository interface {
	Create(ctx context.Context, tx *sql.Tx, userID int64) (int64, error)
	Enroll(ctx context.Context, walletID int, amount decimal.Decimal) (int, error)
	GetByUserId(ctx context.Context, userID int) (*Wallet, error)
	Transfer(ctx context.Context, walletFrom, walletTo int, amount decimal.Decimal) (int, error)
}

type WalletService struct {
	db *sql.DB
}

func NewWalletService(db *sql.DB) SQLRepository {
	return WalletService{
		db: db,
	}
}

// Create creates new wallet for user
func (ws WalletService) Create(ctx context.Context, tx *sql.Tx, userID int64) (int64, error) {
	// tx, txErr := conn.BeginTx(ctx, nil)
	// if txErr != nil {
	// 	return 0, fmt.Errorf("Error of transaction initialization: %s", txErr)
	// }
	// tx.ExecContext(ctx, "set transaction isolation level serializable")
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

	// txCommitErr := tx.Commit()
	// if txCommitErr != nil {
	// 	tx.Rollback()
	// 	return 0, fmt.Errorf("Error of transaction commit: %s", txCommitErr)
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
	conn, _ := ws.db.Conn(ctx)

	// Apply AdvisoryLock for operation
	_, alErr := conn.ExecContext(ctx, "select pg_advisory_lock(?)", EnrollWallet)
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
	_, updateErr := tx.ExecContext(ctx, "update wallets set balance=balance+? where id=?", amount, walletID)
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

// GetByUserId retrieves wallet by user ID
func (ws WalletService) GetByUserId(ctx context.Context, userID int) (*Wallet, error) {
	wallet := Wallet{}
	getWalletErr := ws.db.
		QueryRowContext(ctx, "select id, user_id, balance, currency from wallets where user_id=?", userID).
		Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.Currency)
	if getWalletErr != nil {
		return nil, getWalletErr
	}
	return &wallet, nil
}

// Transfer moves financial resources from one wallet to another
func (ws WalletService) Transfer(ctx context.Context, walletFrom, walletTo int, amount decimal.Decimal) (int, error) {

	// Receive source wallet
	sourceWallet := Wallet{}
	getSourceWalletErr := ws.db.
		QueryRowContext(ctx, "select id, user_id, balance, currency from wallets where id=?", walletFrom).
		Scan(&sourceWallet.ID, &sourceWallet.UserID, &sourceWallet.Balance, &sourceWallet.Currency)
	if getSourceWalletErr != nil {
		return 0, fmt.Errorf("error of receiving source wallet data: %s", getSourceWalletErr)
	}

	if sourceWallet.Balance.LessThanOrEqual(decimal.Zero) {
		return 0, fmt.Errorf("source wallet balance is less or equal to zero")
	}

	// Get connection from the pool
	conn, _ := ws.db.Conn(ctx)

	// Apply AdvisoryLock for operation
	_, alErr := conn.ExecContext(ctx, "select pg_advisory_lock(?)", TransferFunds)
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
	_, updateSourceErr := tx.ExecContext(ctx, "update wallets set balance=balance-? where id=?", amount, walletFrom)
	if updateSourceErr != nil {
		tx.Rollback()
		conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, TransferFunds)
		return 0, fmt.Errorf("error source wallet debit: %s", updateSourceErr)
	}

	// Update target wallet 'balance' column
	_, updateTargetErr := tx.ExecContext(ctx, "update wallets set balance=balance+? where id=?", amount, walletTo)
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
