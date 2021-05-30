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
)

type SQLRepository interface {
	Create(ctx context.Context, conn *sql.Conn, userID int64) (int64, error)
	Enroll(ctx context.Context, walletID int, amount decimal.Decimal) (int, error)
	GetByUserId(ctx context.Context, userID int) (*Wallet, error)
	// Transfer(walletFrom, walletTo int, amount decimal.Decimal) (*Wallet, error)
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
func (ws WalletService) Create(ctx context.Context, conn *sql.Conn, userID int64) (int64, error) {
	tx, txErr := conn.BeginTx(ctx, nil)
	if txErr != nil {
		return 0, fmt.Errorf("Error of transaction initialization: %s", txErr)
	}
	tx.ExecContext(ctx, "set transaction isolation level serializable")

	insertRes, insertErr := tx.ExecContext(
		ctx,
		"insert into wallets(user_id) values(?)",
		userID,
	)

	if insertErr != nil {
		return 0, fmt.Errorf("Error wallet creation: %s", insertErr)
	}

	txCommitErr := tx.Commit()
	if txCommitErr != nil {
		tx.Rollback()
		return 0, fmt.Errorf("Error of transaction commit: %s", txCommitErr)
	}

	return insertRes.LastInsertId()
}

// Enroll updates wallet's balance
func (ws WalletService) Enroll(ctx context.Context, walletID int, amount decimal.Decimal) (int, error) {
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
