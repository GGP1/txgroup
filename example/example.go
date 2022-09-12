package example_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/GGP1/txgroup"
)

// User is the request body schema for this example.
type User struct {
	ID string `json:"id,omitempty"`
}

type handler struct {
	service service
	db      *sql.DB
}

// Error handling omitted for simplicity sake.
func (h handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var user User
	_ = json.NewDecoder(r.Body).Decode(&user)
	defer r.Body.Close()

	// Create sql transaction, you'd do the same for other databases
	sqlTx, _ := NewSQLTx(ctx, h.db, 1)

	txg, ctx := txgroup.WithContext(ctx, sqlTx)
	defer txg.Rollback()

	_ = h.service.CreateUser(ctx, user)

	if err := txg.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write response
}

type service struct{}

func (s service) CreateUser(ctx context.Context, user User) error {
	sqlTx := SQLTxFromContext(ctx)
	q := "INSERT INTO users (id) VALUES ($1)"
	if _, err := sqlTx.ExecContext(ctx, q, user.ID); err != nil {
		return err
	}
	// Perform executions in other databases ...
	return nil
}
