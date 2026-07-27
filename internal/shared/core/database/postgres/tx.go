package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// TxFunc is the function executed inside a transaction.
// Returning any error triggers an automatic rollback.
type TxFunc func(ctx context.Context, tx pgx.Tx) error

// WithTx runs fn inside a single database transaction.
//
// Commit is called if fn returns nil.
// Rollback is called if fn returns an error or panics.
//
// Usage:
//
//	err := db.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
//	    if _, err := tx.Exec(ctx, "INSERT INTO users ..."); err != nil {
//	        return err   // triggers rollback
//	    }
//	    if _, err := tx.Exec(ctx, "INSERT INTO profiles ..."); err != nil {
//	        return err   // triggers rollback
//	    }
//	    return nil       // triggers commit
//	})
func (db *DB) WithTx(ctx context.Context, fn TxFunc) (err error) {
	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Ensure rollback on panic or error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // re-panic after rollback
		}
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if err = fn(ctx, tx); err != nil {
		return err // deferred rollback fires
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// WithTxSerializable runs fn inside a SERIALIZABLE transaction.
// Use this for operations that require strict isolation (e.g. balance transfers,
// inventory deductions, any read-then-write pattern that must be atomic).
func (db *DB) WithTxSerializable(ctx context.Context, fn TxFunc) error {
	return db.withTxIsolation(ctx, pgx.Serializable, fn)
}

func (db *DB) withTxIsolation(ctx context.Context, level pgx.TxIsoLevel, fn TxFunc) (err error) {
	tx, err := db.Pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: level,
	})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	if err = fn(ctx, tx); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
