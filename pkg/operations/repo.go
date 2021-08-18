package operations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
)

const (
	Retrieve   = "retrieve"
	Create     = "create wallet"
	Deposit    = "deposit"
	Withdrawal = "withdrawal"
)

type IWalletOperationRepo interface {
	Create(ctx context.Context, tx *sql.Tx, operation string, walletFrom, walletTo int, amount decimal.Decimal) (int, error)
	List(ctx context.Context) (*sql.Rows, error)
}

type WalletOperationService struct {
	db *sql.DB
}

func NewWalletOperationRepo(db *sql.DB) IWalletOperationRepo {
	return WalletOperationService{
		db: db,
	}
}

func (wor WalletOperationService) Create(ctx context.Context, tx *sql.Tx, operation string, walletFrom, walletTo int, amount decimal.Decimal) (int, error) {
	var (
		walletOperationID int
		walletFromValue   interface{}
	)

	if walletFrom == 0 {
		walletFromValue = nil
	} else {
		walletFromValue = walletFrom
	}

	stmt, insertErr := tx.QueryContext(
		ctx,
		"insert into wallet_operations(operation, wallet_from, wallet_to, amount) values($1, $2, $3, $4) returning id",
		operation, walletFromValue, walletTo, amount,
	)

	if insertErr != nil {
		return 0, fmt.Errorf("error wallet operation creation: %s", insertErr)
	}

	for stmt.Next() {
		scanErr := stmt.Scan(&walletOperationID)
		if scanErr != nil {
			return 0, fmt.Errorf("error wallet operation id retrieving: %s", insertErr)
		}
	}

	return walletOperationID, nil
}

func (wor WalletOperationService) List(ctx context.Context) (*sql.Rows, error) {
	return wor.db.QueryContext(ctx, "select id, operation, wallet_from, wallet_to, amount, created_at from wallet_operations")
}
