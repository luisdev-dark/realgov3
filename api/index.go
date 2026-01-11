package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/luisdev-dark/realgov3.git/db"
	"github.com/luisdev-dark/realgov3.git/routes"
)

var (
	router     http.Handler
	routerOnce sync.Once
)

// Handler es el entrypoint que Vercel usa para esta Function.
func Handler(w http.ResponseWriter, r *http.Request) {
	routerOnce.Do(func() {
		// Inicializar DB
		if err := db.InitDB(); err != nil {
			log.Printf("Error inicializando DB: %v", err)
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
			return
		}
		// Configurar router
		r := routes.SetupRouter()
		// Vercel envía la ruta completa (ej: /api/routes), pero el router espera /routes
		// Usamos StripPrefix para remover /api
		router = http.StripPrefix("/api", r)
	})

	if router != nil {
		router.ServeHTTP(w, r)
	}
}
