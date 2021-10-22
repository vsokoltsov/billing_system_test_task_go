package repositories

import (
	"billing_system_test_task/internal/adapters/tx"
	"billing_system_test_task/internal/entities"
	"context"
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
	WithTx(t tx.Tx) WalletsManager
	Create(ctx context.Context, userID int64) (int64, error)
	Enroll(ctx context.Context, walletID int, amount decimal.Decimal) (int, error)
	GetByUserId(ctx context.Context, userID int) (*entities.Wallet, error)
	GetByID(ctx context.Context, walletID int) (*entities.Wallet, error)
	Transfer(ctx context.Context, walletFrom, walletTo int, amount decimal.Decimal) (int, error)
}

// WalletService shows structure for service of wallets
type WalletService struct {
	db tx.SQLQueryAdapter
}

// NewWalletService returns instance of WalletService
func NewWalletService(db tx.SQLQueryAdapter) *WalletService {
	return &WalletService{
		db: db,
	}
}

func (ws WalletService) WithTx(t tx.Tx) WalletsManager {
	return NewWalletService(t.(tx.SQLQueryAdapter))
}

// Create creates new wallet for user
func (ws WalletService) Create(ctx context.Context, userID int64) (int64, error) {
	var (
		walletID int64
	)

	stmt, insertErr := ws.db.QueryContext(
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

	return walletID, nil
}

// Enroll updates wallet's balance
func (ws WalletService) Enroll(ctx context.Context, walletID int, amount decimal.Decimal) (int, error) {
	// Check if amount is less or equal to 0
	if amount.LessThanOrEqual(decimal.Zero) {
		return 0, fmt.Errorf("amount should be greater than 0")
	}

	// Update wallet 'balance' column
	_, updateErr := ws.db.ExecContext(ctx, "update wallets set balance=balance+$1 where id=$2 returning balance", amount, walletID)
	if updateErr != nil {
		// _ = tx.Rollback()
		// _, _ = conn.ExecContext(ctx, `select pg_advisory_unlock($1)`, EnrollWallet)
		return 0, fmt.Errorf("error wallet enrollment: %s", updateErr)
	}

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
	// Update source wallet 'balance' column
	_, updateSourceErr := ws.db.ExecContext(ctx, "update wallets set balance=balance-$1 where id=$2", amount, walletFrom)
	if updateSourceErr != nil {
		return 0, fmt.Errorf("error source wallet debit: %s", updateSourceErr)
	}

	// Update target wallet 'balance' column
	_, updateTargetErr := ws.db.ExecContext(ctx, "update wallets set balance=balance+$1 where id=$2", amount, walletTo)
	if updateTargetErr != nil {
		return 0, fmt.Errorf("error target wallet transfer: %s", updateTargetErr)
	}

	return walletFrom, nil
}
