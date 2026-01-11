package db

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DATABASE_URL debe incluir sslmode=require,
// por ejemplo: postgres://user:pass@host/dbname?sslmode=require

var (
	pool    *pgxpool.Pool
	poolErr error
	once    sync.Once
)

// GetPool inicializa (lazy) y retorna un pool global reutilizable.
func GetPool(ctx context.Context) (*pgxpool.Pool, error) {
	once.Do(func() {
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			poolErr = fmt.Errorf("DATABASE_URL no está definida")
			return
		}

		p, err := pgxpool.New(ctx, dsn)
		if err != nil {
			poolErr = err
			return
		}

		if err := p.Ping(ctx); err != nil {
			p.Close()
			poolErr = err
			return
		}

		pool = p
	})

	return pool, poolErr
}