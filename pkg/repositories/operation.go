package repositories

import (
	"billing_system_test_task/pkg/adapters/tx"
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

type OperationsManager interface {
	WithTx(t tx.Tx) OperationsManager
	Create(ctx context.Context, tx *sql.Tx, operation string, walletFrom, walletTo int, amount decimal.Decimal) (int, error)
	List(ctx context.Context, params *ListParams) (*sql.Rows, error)
}

type WalletOperationService struct {
	db tx.SQLQueryAdapter
}

type ListParams struct {
	Page    int
	PerPage int
	Date    string
}

func NewWalletOperationRepo(db tx.SQLQueryAdapter) OperationsManager {
	return &WalletOperationService{
		db: db,
	}
}

func (wor WalletOperationService) WithTx(t tx.Tx) OperationsManager {
	return NewWalletOperationRepo(t.(tx.SQLQueryAdapter))
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
			return 0, fmt.Errorf("error wallet operation id retrieving: %s", scanErr)
		}
	}

	return walletOperationID, nil
}

func (wor WalletOperationService) List(ctx context.Context, params *ListParams) (*sql.Rows, error) {
	query := "select id, operation, wallet_from, wallet_to, amount, created_at from wallet_operations"
	if params != nil {
		args := []interface{}{}
		page := params.Page
		if page == 1 {
			page = 0
		} else {
			page -= 1
		}

		if params.Date != "" {
			query += " where created_at = to_date($1, 'YYYY-MM-DD') "
			args = append(args, params.Date)
		}

		if params.PerPage != 0 {
			var (
				pageIdx    int
				perPageIdx int
				argsLen    = len(args)
			)

			if argsLen == 0 {
				pageIdx = 1
				perPageIdx = 2
			} else {
				pageIdx = argsLen + 1
				perPageIdx = pageIdx + 1
			}
			query += fmt.Sprintf(" offset $%d limit $%d", pageIdx, perPageIdx)
			args = append(args, page*params.PerPage)
			args = append(args, params.PerPage)
		}
		fmt.Println(query)
		return wor.db.QueryContext(
			ctx,
			query,
			args...,
		)
	}
	return wor.db.QueryContext(
		ctx,
		query,
	)
}
