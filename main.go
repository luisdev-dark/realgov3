package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/luisdev-dark/realgov3.git/db"
	"github.com/luisdev-dark/realgov3.git/routes"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró archivo .env, usando variables del sistema")
	}

	// Conectar a Postgres
	if err := db.InitDB(); err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}
	defer db.CloseDB()

	// Configurar rutas
	r := routes.SetupRouter()

	// Iniciar servidor
	port := os.Getenv("PORT")
	if port == "" {
		// Render define PORT, este fallback es solo para desarrollo local
		port = "8080"
	}
	addr := ":" + port

	log.Printf("Servidor iniciado en puerto %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
