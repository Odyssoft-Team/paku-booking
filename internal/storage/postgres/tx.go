package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"paku-booking/internal/booking"
)

// Repo implementa booking.Repository usando PostgreSQL.
type Repo struct {
	db *sql.DB
}

// NewRepo crea un repositorio PostgreSQL.
func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

// Tx ejecuta fn dentro de una transacción SQL.
func (r *Repo) Tx(ctx context.Context, fn func(ctx context.Context, tx booking.TxRepo) error) error {
	sqlTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	txRepo := &txRepo{tx: sqlTx}

	err = fn(ctx, txRepo)
	if err != nil {
		if rbErr := sqlTx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed after error %v: %w", err, rbErr)
		}
		return err
	}

	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// txRepo implementa booking.TxRepo usando *sql.Tx.
type txRepo struct {
	tx *sql.Tx
}

// Compile-time interface checks
var _ booking.Repository = (*Repo)(nil)
var _ booking.TxRepo = (*txRepo)(nil)
