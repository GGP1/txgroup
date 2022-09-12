package txgroup_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GGP1/txgroup"
)

func TestNewContext(t *testing.T) {
	ctx := context.Background()

	ctx, cancel := txgroup.NewContext(ctx, txMock, newTxMock("tx", nil, 3))
	defer cancel()

	got, err := txgroup.TxFromContext(ctx, txMock.Key())
	if err != nil {
		t.Error(err)
	}

	if txMock != got {
		t.Errorf("Expected %v, got %v", txMock, got)
	}
}

func TestWithContext(t *testing.T) {
	ctx := context.Background()

	_, ctx = txgroup.WithContext(ctx, txMock)

	got, err := txgroup.TxFromContext(ctx, txMock.Key())
	if err != nil {
		t.Error(err)
	}

	if txMock != got {
		t.Errorf("Expected %v, got %v", txMock, got)
	}
}

func TestCtxKeyCollision(t *testing.T) {
	ctx := context.Background()
	_, ctx = txgroup.WithContext(ctx, txMock)

	if v := ctx.Value("mock_tx"); v != nil {
		t.Errorf("Expected nil and got %v", v)
	}
}

func TestTxFromContext(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		_, ctx = txgroup.WithContext(ctx, txMock)

		tx, err := txgroup.TxFromContext(ctx, txMock.Key())
		if err != nil {
			t.Error(err)
		}

		if txMock.Key() != tx.Key() {
			t.Errorf("Expected %s, got %s", txMock.Key(), tx.Key())
		}
	})

	t.Run("Not found", func(t *testing.T) {
		if _, err := txgroup.TxFromContext(context.Background(), "key"); err == nil {
			t.Error("Expected an error and got nil")
		}
	})

	t.Run("Context canceled", func(t *testing.T) {
		txg, ctx := txgroup.WithContext(nil, txMock)

		if err := txg.Rollback(); err != nil {
			t.Error(err)
		}

		if _, err := txgroup.TxFromContext(ctx, txMock.Key()); err == nil {
			t.Error("Expected an error and got nil")
		}
	})
}

func TestAddTx(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		txg, ctx := txgroup.WithContext(ctx, newTxMock("1", nil, 2), newTxMock("1", nil, 0))

		ctx, err := txg.AddTx(ctx, txMock)
		if err != nil {
			t.Error(err)
		}

		tx, err := txgroup.TxFromContext(ctx, txMock.Key())
		if err != nil {
			t.Error(err)
		}

		if txMock.Key() != tx.Key() {
			t.Errorf("Expected %s, got %s", txMock.Key(), tx.Key())
		}
	})

	t.Run("Context canceled", func(t *testing.T) {
		txg, ctx := txgroup.WithContext(nil)

		ctx, err := txg.AddTx(ctx, txMock)
		if err != nil {
			t.Error(err)
		}

		if err := txg.Rollback(); err != nil {
			t.Error(err)
		}

		if _, err := txg.AddTx(ctx, newTxMock("ctx_canceled", nil, 2)); err == nil {
			t.Error("Expected an error and got nil")
		}
	})
}

func TestCommit(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		txg, ctx := txgroup.WithContext(ctx, txMock)

		if err := txg.Commit(); err != nil {
			t.Error(err)
		}
	})

	t.Run("Fail", func(t *testing.T) {
		ctx := context.Background()
		txg, ctx := txgroup.WithContext(ctx, txMockErr)

		if err := txg.Commit(); err == nil {
			t.Error("Expected an error and got nil")
		}
	})
}

func TestRollback(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		txg, ctx := txgroup.WithContext(ctx, txMock)

		if err := txg.Rollback(); err != nil {
			t.Error(err)
		}
	})

	t.Run("Fail", func(t *testing.T) {
		ctx := context.Background()
		txg, ctx := txgroup.WithContext(ctx, txMockErr)

		if err := txg.Rollback(); err == nil {
			t.Error("Expected an error and got nil")
		}
	})
}

var (
	txMock    = newTxMock("tx_mock", nil, 1)
	txMockErr = newTxMock("tx_mock_err", errors.New("error"), 0)
)

func newTxMock(key string, err error, weight uint) *tx {
	return &tx{
		key:    key,
		err:    err,
		weight: weight,
	}
}

type tx struct {
	err    error
	key    string
	weight uint
}

func (t *tx) Commit() error {
	return t.err
}

func (t *tx) Key() string {
	return t.key
}

func (t *tx) Rollback() error {
	return t.err
}

func (t *tx) Weight() uint {
	return t.weight
}
