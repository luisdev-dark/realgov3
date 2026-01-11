package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/luisdev-dark/realgov3.git/db"
	"github.com/luisdev-dark/realgov3.git/models"
)

// GetRoutes retorna todas las rutas activas
//
// Request:
// GET /routes
//
// Response:
// 200 OK
// [
//   {
//     "id": "uuid",
//     "name": "Ruta Centro - Norte",
//     "origin_name": "Centro",
//     "destination_name": "Norte",
//     "base_price_cents": 500,
//     "currency": "PEN",
//     "is_active": true
//   }
// ]
func GetRoutes(w http.ResponseWriter, r *http.Request) {
	pool := db.GetDB()

	query := `
		SELECT id, name, is_active, origin_name, origin_lat, origin_lon,
		       destination_name, destination_lat, destination_lon,
		       base_price_cents, currency, created_at, updated_at
		FROM app.routes
		WHERE is_active = true
		ORDER BY created_at DESC
	`

	rows, err := pool.Query(r.Context(), query)
	if err != nil {
		http.Error(w, "Error consultando rutas", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var routes []models.Route
	for rows.Next() {
		var route models.Route
		if err := rows.Scan(
			&route.ID,
			&route.Name,
			&route.IsActive,
			&route.OriginName,
			&route.OriginLat,
			&route.OriginLon,
			&route.DestinationName,
			&route.DestinationLat,
			&route.DestinationLon,
			&route.BasePriceCents,
			&route.Currency,
			&route.CreatedAt,
			&route.UpdatedAt,
		); err != nil {
			http.Error(w, "Error escaneando rutas", http.StatusInternalServerError)
			return
		}
		routes = append(routes, route)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routes)
}

// GetRouteByID retorna el detalle de una ruta con sus paradas
//
// Request:
// GET /routes/{id}
//
// Response:
// 200 OK
// {
//   "id": "uuid",
//   "name": "Ruta Centro - Norte",
//   "origin": "Centro",
//   "destination": "Norte",
//   "base_price": 5.00,
//   "stops": [
//     {"id": "uuid1", "name": "Parada A"},
//     {"id": "uuid2", "name": "Parada B"}
//   ]
// }
func GetRouteByID(w http.ResponseWriter, r *http.Request) {
	pool := db.GetDB()

	routeID := chi.URLParam(r, "id")
	if routeID == "" {
		http.Error(w, "ID de ruta requerido", http.StatusBadRequest)
		return
	}

	// Validar UUID
	_, err := uuid.Parse(routeID)
	if err != nil {
		http.Error(w, "ID de ruta inválido", http.StatusBadRequest)
		return
	}

	// Consultar ruta
	routeQuery := `
		SELECT id, name, is_active, origin_name, origin_lat, origin_lon,
		       destination_name, destination_lat, destination_lon,
		       base_price_cents, currency, created_at, updated_at
		FROM app.routes
		WHERE id = $1
	`

	var route models.Route
	err = pool.QueryRow(r.Context(), routeQuery, routeID).Scan(
		&route.ID,
		&route.Name,
		&route.IsActive,
		&route.OriginName,
		&route.OriginLat,
		&route.OriginLon,
		&route.DestinationName,
		&route.DestinationLat,
		&route.DestinationLon,
		&route.BasePriceCents,
		&route.Currency,
		&route.CreatedAt,
		&route.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Ruta no encontrada", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Error consultando ruta", http.StatusInternalServerError)
		return
	}

	// Consultar paradas de la ruta
	stopsQuery := `
		SELECT id, name
		FROM app.route_stops
		WHERE route_id = $1 AND is_active = true
		ORDER BY stop_order ASC
	`

	rows, err := pool.Query(r.Context(), stopsQuery, routeID)
	if err != nil {
		http.Error(w, "Error consultando paradas", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stops []models.StopInfo
	for rows.Next() {
		var stop models.StopInfo
		if err := rows.Scan(&stop.ID, &stop.Name); err != nil {
			http.Error(w, "Error escaneando paradas", http.StatusInternalServerError)
			return
		}
		stops = append(stops, stop)
	}

	// Construir respuesta
	basePrice := float64(route.BasePriceCents) / 100.0
	routeDetail := models.RouteDetail{
		ID:          route.ID,
		Name:        route.Name,
		Origin:      route.OriginName,
		Destination: route.DestinationName,
		BasePrice:   basePrice,
		Stops:       stops,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routeDetail)
}
