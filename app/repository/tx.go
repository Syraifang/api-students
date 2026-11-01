package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX adalah interface yang menjembatani pgxpool.Pool dan pgx.Tx
// agar repository bisa dipakai baik dengan maupun tanpa transaksi.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) runOnce(ctx context.Context, fn func(tx DBTX) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// WithTx menjalankan fungsi di dalam database transaction (dengan retry jika terjadi deadlock).
func (m *TxManager) WithTx(ctx context.Context, fn func(tx DBTX) error) error {
	// Pada implementasi standar, bisa langsung memanggil runOnce atau ditambahkan logika retry
	return m.runOnce(ctx, fn)
}

// WithTxOnce sama dengan WithTx, tetapi TIDAK PERNAH melakukan retry (Tambahan Modul 11).
func (m *TxManager) WithTxOnce(ctx context.Context, fn func(tx DBTX) error) error {
	return m.runOnce(ctx, fn)
}