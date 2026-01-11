package main

import (
	"log"
	"net/http"

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
	port := ":8080"
	log.Printf("Servidor iniciado en puerto %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}
