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
	// Cargar .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error cargando .env:", err)
	}

	// Obtener DATABASE_URL
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL no está definida en .env")
	}

	fmt.Println("Conectando a Neon...")
	fmt.Println("DATABASE_URL:", connStr[:50]+"...") // Solo muestra primeros 50 chars

	// Conectar
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatal("Error conectando a Neon:", err)
	}
	defer conn.Close(context.Background())

	fmt.Println("✅ Conexión exitosa a Neon!")

	// Verificar versión de Postgres
	var version string
	err = conn.QueryRow(context.Background(), "SELECT version()").Scan(&version)
	if err != nil {
		log.Fatal("Error consultando versión:", err)
	}
	fmt.Println("📦 Versión de Postgres:", version[:50]+"...")

	// Verificar tablas existentes
	fmt.Println("\n📋 Tablas en la base de datos:")
	rows, err := conn.Query(context.Background(), `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'app'
		ORDER BY table_name
	`)
	if err != nil {
		log.Fatal("Error consultando tablas:", err)
	}
	defer rows.Close()

	tableCount := 0
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatal("Error escaneando tabla:", err)
		}
		fmt.Printf("  - %s\n", tableName)
		tableCount++
	}

	if tableCount == 0 {
		fmt.Println("  ⚠️  No hay tablas en el schema 'app'. Ejecuta seed.sql primero.")
	} else {
		fmt.Printf("  Total: %d tablas\n", tableCount)
	}

	// Verificar rutas
	fmt.Println("\n🚀 Rutas en la base de datos:")
	routeRows, err := conn.Query(context.Background(), "SELECT id, name, origin, destination FROM app.routes")
	if err != nil {
		fmt.Println("  ⚠️  No hay rutas o tabla no existe")
	} else {
		defer routeRows.Close()
		for routeRows.Next() {
			var id, name, origin, destination string
			if err := routeRows.Scan(&id, &name, &origin, &destination); err != nil {
				log.Fatal("Error escaneando ruta:", err)
			}
			fmt.Printf("  - %s: %s → %s\n", name, origin, destination)
		}
	}

	fmt.Println("\n✅ Prueba de conexión completada!")
}
