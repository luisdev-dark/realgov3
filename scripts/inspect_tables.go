package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error cargando .env:", err)
	}

	connStr := os.Getenv("DATABASE_URL")
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("Error conectando:", err)
	}
	defer conn.Close(context.Background())

	// Inspeccionar estructura de cada tabla
	tables := []string{"users", "routes", "route_stops", "trips"}

	for _, table := range tables {
		fmt.Printf("\n📋 Tabla: app.%s\n", table)
		fmt.Println("─────────────────────────────────")

		query := `
			SELECT column_name, data_type, is_nullable
			FROM information_schema.columns
			WHERE table_schema = 'app' AND table_name = $1
			ORDER BY ordinal_position
		`

		rows, err := conn.Query(context.Background(), query, table)
		if err != nil {
			fmt.Printf("  ⚠️  Error: %v\n", err)
			continue
		}
		defer rows.Close()

		hasColumns := false
		for rows.Next() {
			hasColumns = true
			var colName, dataType, isNullable string
			if err := rows.Scan(&colName, &dataType, &isNullable); err != nil {
				log.Fatal("Error escaneando:", err)
			}
			fmt.Printf("  - %-20s %-15s (nullable: %s)\n", colName, dataType, isNullable)
		}

		if !hasColumns {
			fmt.Println("  ⚠️  Tabla no encontrada")
		}
	}

	// Verificar datos existentes
	fmt.Println("\n\n📊 Datos existentes:")
	fmt.Println("─────────────────────────────────")

	// Users
	fmt.Println("\n👥 Users:")
	userRows, _ := conn.Query(context.Background(), "SELECT * FROM app.users LIMIT 5")
	if userRows != nil {
		defer userRows.Close()
		for userRows.Next() {
			var id string
			var email string
			userRows.Scan(&id, &email)
			fmt.Printf("  - ID: %s, Email: %s\n", id, email)
		}
	}

	// Routes
	fmt.Println("\n🚀 Routes:")
	routeRows, _ := conn.Query(context.Background(), "SELECT id, name, origin, destination FROM app.routes LIMIT 5")
	if routeRows != nil {
		defer routeRows.Close()
		for routeRows.Next() {
			var id, name, origin, destination string
			routeRows.Scan(&id, &name, &origin, &destination)
			fmt.Printf("  - %s: %s → %s\n", name, origin, destination)
		}
	}
}
