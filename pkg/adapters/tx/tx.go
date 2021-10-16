package tx

import (
	"context"
	"database/sql"
	"fmt"
)

type SQLAdapter interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	SQLQueryAdapter
}

type SQLQueryAdapter interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type Tx interface {
	Commit() error
	Rollback() error
}

type TxBeginner interface {
	BeginTrx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
}

type txBeginner struct {
	SQLAdapter
}

func NewTxBeginner(sqlDB SQLAdapter) TxBeginner {
	return &txBeginner{sqlDB}
}

func (tb *txBeginner) BeginTrx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {

	sqlTx, txErr := tb.BeginTx(ctx, opts)
	if txErr != nil {
		return nil, fmt.Errorf("transaction initialization error: %s", txErr)
	}
	return &tx{sqlTx}, nil
}

type tx struct {
	*sql.Tx
}

func (t *tx) Commit() error {
	return t.Tx.Commit()
}

func (t *tx) Rollback() error {
	return t.Tx.Rollback()
}
