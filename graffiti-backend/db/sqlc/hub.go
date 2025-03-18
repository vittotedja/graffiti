package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Hub provides all functions to execute db queries and transactions
type Hub struct {
	// a composition, a preferred way to extend struct functionality in Golang instead of inheritance
	// All individual query functions are defined in the Queries struct
	*Queries
	pool *pgxpool.Pool
}

func NewHub(pool *pgxpool.Pool) *Hub {
	return &Hub{
		pool:    pool,
		Queries: New(pool),
	}
}

// execTx executes a function within a database transaction
// It rolls back the transaction if the function returns an error
func (hub *Hub) execTx(ctx context.Context, fn func(*Queries) error) error {
	// Create empty TxOptions for default options
	txOptions := pgx.TxOptions{}
	
	tx, err := hub.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}
	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit(ctx)
}