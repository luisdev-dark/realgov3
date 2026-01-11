package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/luisdev-dark/realgov3.git/handlers"
)

// SetupRouter configura las rutas del MVP
func SetupRouter() *chi.Mux {
	r := chi.NewRouter()

	// Rutas de rutas (routes)
	r.Get("/routes", handlers.GetRoutes)
	r.Get("/routes/{id}", handlers.GetRouteByID)

	// Rutas de viajes (trips)
	r.Post("/trips", handlers.CreateTrip)
	r.Get("/trips/{id}", handlers.GetTripByID)

	return r
}
