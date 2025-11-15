package repository

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Пул или транзакция
type DB interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txKey struct{}

// Транзакция
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type PgTxManager struct {
	db *pgxpool.Pool
}

func NewPgTxManager(db *pgxpool.Pool) *PgTxManager {
	return &PgTxManager{db: db}
}

func (m *PgTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Printf("tx rollback error: %v", err)
		}
	}()

	ctxWithTx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(ctxWithTx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// currentDB — отдает либо открытую транзакцию, либо пул.
func currentDB(ctx context.Context, pool *pgxpool.Pool) DB {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return pool
}
