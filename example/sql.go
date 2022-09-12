package example_test

import (
	"context"
	"database/sql"

	"github.com/GGP1/txgroup"
)

const sqlKey = "sql"

// SQL represents an sql transaction.
type SQL struct {
	tx     *sql.Tx
	weight uint
}

// NewSQLTx returns a new SQL transaction wrapped by a structure
// that satisfies the txgroup.Tx interface.
func NewSQLTx(ctx context.Context, db *sql.DB, weight uint) (*SQL, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &SQL{
		tx:     tx,
		weight: weight,
	}, nil
}

// SQLTxFromContext returns an sql transaction from the context.
func SQLTxFromContext(ctx context.Context) *sql.Tx {
	tx, err := txgroup.TxFromContext(ctx, sqlKey)
	if err != nil {
		panic(err)
	}
	sql, ok := tx.(*SQL)
	if !ok {
		panic("transaction is not of type sql")
	}
	return sql.tx
}

// Key returns the key used for storing SQL transactions in a context.
func (s *SQL) Key() string {
	return sqlKey
}

// Commit commits the transaction.
func (s *SQL) Commit() error {
	return s.tx.Commit()
}

// Rollback discards the transaction.
func (s *SQL) Rollback() error {
	return s.tx.Rollback()
}

// Weight returns the priority number of the transaction.
func (s *SQL) Weight() uint {
	return s.weight
}
