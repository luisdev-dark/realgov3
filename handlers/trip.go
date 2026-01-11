package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/luisdev-dark/realgov3.git/db"
	"github.com/luisdev-dark/realgov3.git/models"
)

// UUID dummy para el usuario (TODO: reemplazar con autenticación real)
const dummyUserID = "00000000-0000-0000-0000-000000000001"

// CreateTripRequest estructura para crear un viaje
type CreateTripRequest struct {
	RouteID        uuid.UUID  `json:"route_id"`
	PickupStopID   *uuid.UUID `json:"pickup_stop_id"`
	DropoffStopID  *uuid.UUID `json:"dropoff_stop_id"`
	PaymentMethod  string      `json:"payment_method"` // cash, yape, pling
}

// CreateTrip crea un nuevo viaje
//
// Request:
// POST /trips
// {
//   "route_id": "uuid-de-la-ruta",
//   "pickup_stop_id": "uuid-parada-recogida | null",
//   "dropoff_stop_id": "uuid-parada-dejada | null",
//   "payment_method": "cash"
// }
//
// Response:
// 200 OK
// {
//   "id": "uuid-del-viaje",
//   "route_id": "uuid-de-la-ruta",
//   "passenger_id": "00000000-0000-0000-0000-000000000001",
//   "pickup_stop_id": "uuid-parada-recogida | null",
//   "dropoff_stop_id": "uuid-parada-dejada | null",
//   "status": "requested",
//   "payment_method": "cash",
//   "price_cents": 500,
//   "currency": "PEN",
//   "scheduled_at": "2026-01-10T10:00:00Z",
//   "created_at": "2026-01-09T15:30:00Z",
//   "updated_at": "2026-01-09T15:30:00Z"
// }
func CreateTrip(w http.ResponseWriter, r *http.Request) {
	pool := db.GetDB()

	var req CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error decodificando request", http.StatusBadRequest)
		return
	}

	// Validar campos requeridos
	if req.RouteID == uuid.Nil {
		http.Error(w, "route_id es requerido", http.StatusBadRequest)
		return
	}
	if req.PaymentMethod == "" {
		http.Error(w, "payment_method es requerido", http.StatusBadRequest)
		return
	}

	// Validar payment_method
	validMethods := map[string]bool{"cash": true, "yape": true, "pling": true}
	if !validMethods[req.PaymentMethod] {
		http.Error(w, "payment_method inválido (cash, yape, pling)", http.StatusBadRequest)
		return
	}

	// Verificar que la ruta existe y obtener precio
	var routeExists bool
	var basePriceCents int
	var currency string
	err := pool.QueryRow(r.Context(),
		"SELECT EXISTS(SELECT 1 FROM app.routes WHERE id = $1), base_price_cents, currency FROM app.routes WHERE id = $1",
		req.RouteID).Scan(&routeExists, &basePriceCents, &currency)
	if err != nil || !routeExists {
		http.Error(w, "Ruta no encontrada", http.StatusNotFound)
		return
	}

	// Si se proporcionan paradas, verificar que existen y pertenecen a la ruta
	if req.PickupStopID != nil {
		var stopExists bool
		err := pool.QueryRow(r.Context(),
			"SELECT EXISTS(SELECT 1 FROM app.route_stops WHERE id = $1 AND route_id = $2)",
			*req.PickupStopID, req.RouteID).Scan(&stopExists)
		if err != nil || !stopExists {
			http.Error(w, "Parada de recogida no encontrada en esta ruta", http.StatusBadRequest)
			return
		}
	}

	if req.DropoffStopID != nil {
		var stopExists bool
		err := pool.QueryRow(r.Context(),
			"SELECT EXISTS(SELECT 1 FROM app.route_stops WHERE id = $1 AND route_id = $2)",
			*req.DropoffStopID, req.RouteID).Scan(&stopExists)
		if err != nil || !stopExists {
			http.Error(w, "Parada de dejada no encontrada en esta ruta", http.StatusBadRequest)
			return
		}
	}

	// Crear el viaje
	tripID := uuid.New()
	passengerID := uuid.MustParse(dummyUserID)
	now := time.Now()
	scheduledAt := now.Add(24 * time.Hour) // Programar para mañana por defecto

	query := `
		INSERT INTO app.trips (id, route_id, passenger_id, pickup_stop_id, dropoff_stop_id, status, payment_method, price_cents, currency, scheduled_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'requested', $6, $7, $8, $9, $10, $10)
		RETURNING id, route_id, passenger_id, pickup_stop_id, dropoff_stop_id, status, payment_method, price_cents, currency, scheduled_at, created_at, updated_at
	`

	var trip models.Trip
	err = pool.QueryRow(r.Context(), query,
		tripID,
		req.RouteID,
		passengerID,
		req.PickupStopID,
		req.DropoffStopID,
		req.PaymentMethod,
		basePriceCents,
		currency,
		scheduledAt,
		now,
	).Scan(
		&trip.ID,
		&trip.RouteID,
		&trip.PassengerID,
		&trip.PickupStopID,
		&trip.DropoffStopID,
		&trip.Status,
		&trip.PaymentMethod,
		&trip.PriceCents,
		&trip.Currency,
		&trip.ScheduledAt,
		&trip.CreatedAt,
		&trip.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "Error creando viaje", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(trip)
}

// GetTripByID retorna el estado completo de un viaje
//
// Request:
// GET /trips/{id}
//
// Response:
// 200 OK
// {
//   "id": "uuid-del-viaje",
//   "passenger_id": "00000000-0000-0000-0000-000000000001",
//   "route": {
//     "id": "uuid-de-la-ruta",
//     "name": "Ruta Centro - Norte",
//     "origin": "Centro",
//     "destination": "Norte",
//     "base_price": 5.00
//   },
//   "pickup": {
//     "id": "uuid-parada",
//     "name": "Parada A"
//   },
//   "dropoff": {
//     "id": "uuid-parada",
//     "name": "Parada B"
//   },
//   "status": "requested",
//   "payment_method": "cash",
//   "price": 5.00,
//   "currency": "PEN",
//   "scheduled_at": "2026-01-10T10:00:00Z",
//   "created_at": "2026-01-09T15:30:00Z"
// }
func GetTripByID(w http.ResponseWriter, r *http.Request) {
	pool := db.GetDB()

	tripID := chi.URLParam(r, "id")
	if tripID == "" {
		http.Error(w, "ID de viaje requerido", http.StatusBadRequest)
		return
	}

	// Validar UUID
	_, err := uuid.Parse(tripID)
	if err != nil {
		http.Error(w, "ID de viaje inválido", http.StatusBadRequest)
		return
	}

	// Consultar viaje
	tripQuery := `
		SELECT id, route_id, passenger_id, pickup_stop_id, dropoff_stop_id, status, payment_method, price_cents, currency, scheduled_at, created_at, updated_at
		FROM app.trips
		WHERE id = $1
	`

	var trip models.Trip
	err = pool.QueryRow(r.Context(), tripQuery, tripID).Scan(
		&trip.ID,
		&trip.RouteID,
		&trip.PassengerID,
		&trip.PickupStopID,
		&trip.DropoffStopID,
		&trip.Status,
		&trip.PaymentMethod,
		&trip.PriceCents,
		&trip.Currency,
		&trip.ScheduledAt,
		&trip.CreatedAt,
		&trip.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Viaje no encontrado", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Error consultando viaje", http.StatusInternalServerError)
		return
	}

	// Consultar ruta
	routeQuery := `
		SELECT id, name, origin_name, destination_name, base_price_cents
		FROM app.routes
		WHERE id = $1
	`

	var routeID uuid.UUID
	var routeName, originName, destName string
	var basePriceCents int
	err = pool.QueryRow(r.Context(), routeQuery, trip.RouteID).Scan(
		&routeID,
		&routeName,
		&originName,
		&destName,
		&basePriceCents,
	)

	if err != nil {
		http.Error(w, "Error consultando ruta", http.StatusInternalServerError)
		return
	}

	// Consultar paradas (si existen)
	var pickupInfo *models.StopInfo
	if trip.PickupStopID != nil {
		pickupQuery := `SELECT id, name FROM app.route_stops WHERE id = $1`
		var stop models.StopInfo
		err = pool.QueryRow(r.Context(), pickupQuery, *trip.PickupStopID).Scan(&stop.ID, &stop.Name)
		if err == nil {
			pickupInfo = &stop
		}
	}

	var dropoffInfo *models.StopInfo
	if trip.DropoffStopID != nil {
		dropoffQuery := `SELECT id, name FROM app.route_stops WHERE id = $1`
		var stop models.StopInfo
		err = pool.QueryRow(r.Context(), dropoffQuery, *trip.DropoffStopID).Scan(&stop.ID, &stop.Name)
		if err == nil {
			dropoffInfo = &stop
		}
	}

	// Construir respuesta
	price := float64(trip.PriceCents) / 100.0
	tripDetail := models.TripDetail{
		ID:            trip.ID,
		PassengerID:   trip.PassengerID,
		Route: models.RouteInfo{
			ID:          routeID,
			Name:        routeName,
			Origin:      originName,
			Destination: destName,
			BasePrice:   float64(basePriceCents) / 100.0,
		},
		Pickup:        pickupInfo,
		Dropoff:       dropoffInfo,
		Status:        trip.Status,
		PaymentMethod: trip.PaymentMethod,
		Price:         price,
		Currency:      trip.Currency,
		ScheduledAt:   trip.ScheduledAt,
		CreatedAt:     trip.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tripDetail)
}
