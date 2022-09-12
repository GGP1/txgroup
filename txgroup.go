// Package txgroup provides a simple way of handling multiple transactions as if they were
// one single block, taking care of their progration, cancelation and perpetration.
package txgroup

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Tx represents a database transaction.
type Tx interface {
	// Commit commits the transaction.
	Commit() error
	// Rollback aborts the transaction.
	Rollback() error
	// Key returns the transaction identifier,
	// it should be unique among the objects satisfying Tx.
	Key() string
	// Weight returns an integer representation of the priority in which the
	// transaction should be executed.
	Weight() uint
}

// unique type is used to prevent collisions with other context keys.
type unique string

// NewContext returns a new context with the transactions stored in it.
func NewContext(ctx context.Context, txs ...Tx) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	// Avoid allocations when receiving only one transaction
	if len(txs) > 1 {
		sort.Slice(txs, func(i, j int) bool {
			return txs[i].Weight() < txs[j].Weight()
		})
	}
	for _, tx := range txs {
		ctx = context.WithValue(ctx, unique(tx.Key()), tx)
	}
	return context.WithCancel(ctx)
}

// TxFromContext returns a single transaction from the context.
func TxFromContext(ctx context.Context, key string) (Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	tx, ok := ctx.Value(unique(key)).(Tx)
	if !ok {
		return nil, errors.New(key + " transaction not found")
	}
	return tx, nil
}

// Group manages a group of transactions.
type Group struct {
	errOnce sync.Once
	cancel  context.CancelFunc
	txs     []Tx
}

// WithContext returns a transactions group containing a
// context with all the transactions stored in it.
func WithContext(ctx context.Context, txs ...Tx) (*Group, context.Context) {
	ctx, cancel := NewContext(ctx, txs...)
	return &Group{txs: txs, cancel: cancel}, ctx
}

// AddTx adds a new transaction to the group.
func (g *Group) AddTx(ctx context.Context, tx Tx) (context.Context, error) {
	// Prevent the usage of the group after the
	// transactions have been committed or rolled back.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	j := sort.Search(len(g.txs), func(i int) bool {
		return g.txs[i].Weight() >= tx.Weight()
	})
	if j == len(g.txs) {
		g.txs = append(g.txs, tx)
	} else {
		g.txs = append(g.txs[:j+1], g.txs[j:]...)
		g.txs[j] = tx
	}

	return context.WithValue(ctx, unique(tx.Key()), tx), nil
}

// Commit commits all transactions.
func (g *Group) Commit() error {
	for _, tx := range g.txs {
		if err := tx.Commit(); err != nil {
			_ = g.Rollback()
			return fmt.Errorf("%s commit failed: %w", tx.Key(), err)
		}
	}
	if g.cancel != nil {
		g.cancel()
	}
	return nil
}

// Rollback aborts all transactions.
func (g *Group) Rollback() error {
	var err error
	for _, tx := range g.txs {
		if err = tx.Rollback(); err != nil {
			g.errOnce.Do(func() {
				err = fmt.Errorf("%s rollback failed: %w", tx.Key(), err)
			})
		}
	}
	if g.cancel != nil {
		g.cancel()
	}
	return err
}
