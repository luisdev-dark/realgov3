package db

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

// InitDB inicializa el pool de conexiones a Postgres
func InitDB() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL no está definida")
	}

	ctx := context.Background()
	var err error
	pool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		return err
	}

	// Verificar conexión
	if err := pool.Ping(ctx); err != nil {
		return err
	}

	log.Println("Conectado a Postgres")
	return nil
}

// GetDB retorna el pool de conexiones
func GetDB() *pgxpool.Pool {
	return pool
}

// CloseDB cierra el pool de conexiones
func CloseDB() {
	if pool != nil {
		pool.Close()
	}
}
