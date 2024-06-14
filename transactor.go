package transactor

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Transactor interface {
	WithinTransaction(context.Context, func(ctx context.Context) error) error
}

type transactor struct {
	conn *pgxpool.Pool
}

func New(conn *pgxpool.Pool) (*transactor, error) {
	if conn == nil {
		return nil, errors.New("a db connection must be provided")
	}

	return &transactor{conn}, nil
}

type txKey struct{}

// // injectTx injects transaction to context
func InjectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// // extractTx extracts transaction from context
func ExtractTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}

func (t *transactor) WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) error {
	tx, err := t.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		// finalize transaction on panic, etc.
		if errTx := tx.Conn().Close(ctx); errTx != nil {
			log.Printf("close transaction: %v", errTx)
		}
	}()

	// run callback
	err = tFunc(InjectTx(ctx, tx))
	if err != nil {
		// if error, rollback
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			log.Printf("rollback transaction: %v", errRollback)
		}
		return err
	}
	// if no error, commit
	if errCommit := tx.Commit(ctx); errCommit != nil {
		log.Printf("commit transaction: %v", errCommit)
	}
	return nil
}
