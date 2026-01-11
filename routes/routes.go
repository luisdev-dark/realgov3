package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/luisdev-dark/realgov3.git/handlers"
)

// SetupRouter configura las rutas del MVP
func SetupRouter() *chi.Mux {
	r := chi.NewRouter()

	// Middleware CORS simple para Expo / web
	r.Use(corsMiddleware)

	// Healthcheck
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok": true}`))
	})

	// Rutas de rutas (routes)
	r.Get("/routes", handlers.GetRoutes)
	r.Get("/routes/{id}", handlers.GetRouteByID)

	// Rutas de viajes (trips)
	r.Post("/trips", handlers.CreateTrip)
	r.Get("/trips/{id}", handlers.GetTripByID)

	return r
}

// corsMiddleware aplica CORS básico para clientes web (incluido Expo web)
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Responder rápido a preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
