package wallets

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	CreateWallet = iota + 1
)

type SQLRepository interface {
	Create(ctx context.Context, conn *sql.Conn, userID int) (int64, error)
	// Enroll(walletID int, amount decimal.Decimal) (*Wallet, error)
	// GetByUserId(userID int) (*Wallet, error)
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
func (ws WalletService) Create(ctx context.Context, conn *sql.Conn, userID int) (int64, error) {
	// conn, _ := ws.db.Conn(ctx)
	// _, alErr := conn.ExecContext(ctx, `select pg_advisory_lock($1)`, CreateWallet)
	// if alErr != nil {
	// 	return 0, fmt.Errorf("Error of starting advisory lock: %s", alErr)
	// }

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

	// _, auErr := conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, CreateWallet)
	// if auErr != nil {
	// 	return 0, fmt.Errorf(
	// 		"Error of unlocking wallet's %d postgres lock: %s",
	// 		CreateWallet,
	// 		auErr,
	// 	)
	// }
	// conn.Close()

	return insertRes.LastInsertId()
}
