package tx

import (
	"billing_system_test_task/pkg/adapters"
	"context"
	"database/sql"
	"fmt"
)

type Tx interface {
	TxCommit(context context.Context) error
	TxRollback(context context.Context) error
}

type TxBeginner interface {
	Begin(ctx context.Context, opts *sql.TxOptions) (Tx, error)
}

type txBeginner struct {
	adapters.SQLAdapter
}

func NewTxBeginner(sqlDB adapters.SQLAdapter) TxBeginner {
	return txBeginner{sqlDB}
}

func (tb txBeginner) Begin(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	sqlTx, txErr := tb.BeginTx(ctx, opts)
	if txErr != nil {
		return nil, fmt.Errorf("transaction initialization error: %s", txErr)
	}
	return &tx{sqlTx}, nil
}

type tx struct {
	*sql.Tx
}

func (t tx) TxCommit(context context.Context) error {
	return t.Commit()
}

func (t tx) TxRollback(context context.Context) error {
	return t.Rollback()
}
