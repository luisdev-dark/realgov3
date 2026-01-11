package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/luisdev-dark/realgov3.git/api/internal/db"
)

const dummyUserID = "00000000-0000-0000-0000-000000000001"

var (
	router     http.Handler
	routerOnce sync.Once
)

// Handler es el entrypoint que Vercel usa para esta Function.
func Handler(w http.ResponseWriter, r *http.Request) {
	routerOnce.Do(func() {
		router = setupRouter()
	})
	router.ServeHTTP(w, r)
}

func setupRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", healthHandler)
	r.Get("/routes", getRoutesHandler)
	r.Get("/routes/{id}", getRouteByIDHandler)
	r.Post("/trips", createTripHandler)
	r.Get("/trips/{id}", getTripByIDHandler)

	return r
}

// Tipos de dominio

type Route struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	IsActive        bool      `json:"is_active"`
	OriginName      string    `json:"origin_name"`
	OriginLat       float64   `json:"origin_lat"`
	OriginLon       float64   `json:"origin_lon"`
	DestinationName string    `json:"destination_name"`
	DestinationLat  float64   `json:"destination_lat"`
	DestinationLon  float64   `json:"destination_lon"`
	BasePriceCents  int       `json:"base_price_cents"`
	Currency        string    `json:"currency"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type StopInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type RouteDetail struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Origin      string     `json:"origin"`
	Destination string     `json:"destination"`
	BasePrice   float64    `json:"base_price"`
	Stops       []StopInfo `json:"stops"`
}

type Trip struct {
	ID            uuid.UUID  `json:"id"`
	RouteID       uuid.UUID  `json:"route_id"`
	PassengerID   uuid.UUID  `json:"passenger_id"`
	PickupStopID  *uuid.UUID `json:"pickup_stop_id"`
	DropoffStopID *uuid.UUID `json:"dropoff_stop_id"`
	Status        string     `json:"status"`
	PaymentMethod string     `json:"payment_method"`
	PriceCents    int        `json:"price_cents"`
	Currency      string     `json:"currency"`
	ScheduledAt   *time.Time `json:"scheduled_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type RouteInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Origin      string    `json:"origin"`
	Destination string    `json:"destination"`
	BasePrice   float64   `json:"base_price"`
}

type TripDetail struct {
	ID            uuid.UUID  `json:"id"`
	PassengerID   uuid.UUID  `json:"passenger_id"`
	Route         RouteInfo  `json:"route"`
	Pickup        *StopInfo  `json:"pickup"`
	Dropoff       *StopInfo  `json:"dropoff"`
	Status        string     `json:"status"`
	PaymentMethod string     `json:"payment_method"`
	Price         float64    `json:"price"`
	Currency      string     `json:"currency"`
	ScheduledAt   *time.Time `json:"scheduled_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

type CreateTripRequest struct {
	RouteID       uuid.UUID  `json:"route_id"`
	PickupStopID  *uuid.UUID `json:"pickup_stop_id"`
	DropoffStopID *uuid.UUID `json:"dropoff_stop_id"`
	PaymentMethod string     `json:"payment_method"` // cash, yape, pling
}

type errorResponse struct {
	Error string `json:"error"`
}

// Helpers JSON

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// Handlers

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /routes
func getRoutesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pool, err := db.GetPool(ctx)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "error de base de datos")
		return
	}

	const q = `
		SELECT id, name, is_active, origin_name, origin_lat, origin_lon,
		       destination_name, destination_lat, destination_lon,
		       base_price_cents, currency, created_at, updated_at
		FROM app.routes
		WHERE is_active = true
		ORDER BY created_at DESC
	`

	rows, err := pool.Query(ctx, q)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "error consultando rutas")
		return
	}
	defer rows.Close()

	var routes []Route
	for rows.Next() {
		var route Route
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
			writeJSONError(w, http.StatusInternalServerError, "error leyendo rutas")
			return
		}
		routes = append(routes, route)
	}

	writeJSON(w, http.StatusOK, routes)
}

// GET /routes/{id}
func getRouteByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pool, err := db.GetPool(ctx)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "error de base de datos")
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		writeJSONError(w, http.StatusBadRequest, "id de ruta requerido")
		return
	}

	routeID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "id de ruta inválido")
		return
	}

	const routeQuery = `
		SELECT id, name, is_active, origin_name, origin_lat, origin_lon,
		       destination_name, destination_lat, destination_lon,
		       base_price_cents, currency, created_at, updated_at
		FROM app.routes
		WHERE id = $1
	`

	var route Route
	err = pool.QueryRow(ctx, routeQuery, routeID).Scan(
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
	if errors.Is(err, pgx.ErrNoRows) {
		writeJSONError(w, http.StatusNotFound, "ruta no encontrada")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "error consultando ruta")
		return
	}

	const stopsQuery = `
		SELECT id, name
		FROM app.route_stops
		WHERE route_id = $1 AND is_active = true
		ORDER BY stop_order ASC
	`

	rows, err := pool.Query(ctx, stopsQuery, routeID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "error consultando paradas")
		return
	}
	defer rows.Close()

	var stops []StopInfo
	for rows.Next() {
		var stop StopInfo
		if err := rows.Scan(&stop.ID, &stop.Name); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "error leyendo paradas")
			return
		}
		stops = append(stops, stop)
	}

	basePrice := float64(route.BasePriceCents) / 100.0
	resp := RouteDetail{
		ID:          route.ID,
		Name:        route.Name,
		Origin:      route.OriginName,
		Destination: route.DestinationName,
		BasePrice:   basePrice,
		Stops:       stops,
	}

	writeJSON(w, http.StatusOK, resp)
}

// POST /trips
func createTripHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pool, err := db.GetPool(ctx)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "error de base de datos")
		return
	}

	var req CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "json inválido")
		return
	}

	if req.RouteID == uuid.Nil {
		writeJSONError(w, http.StatusBadRequest, "route_id es requerido")
		return
	}
	if req.PaymentMethod == "" {
		writeJSONError(w, http.StatusBadRequest, "payment_method es requerido")
		return
	}

	validMethods := map[string]bool{"cash": true, "yape": true, "pling": true}
	if !validMethods[req.PaymentMethod] {
		writeJSONError(w, http.StatusBadRequest, "payment_method inválido (cash, yape, pling)")
		return
	}

	const routeQuery = `
		SELECT base_price_cents, currency
		FROM app.routes
		WHERE id = $1
	`

	var basePriceCents int
	var currency string
	err = pool.QueryRow(ctx, routeQuery, req.RouteID).Scan(&basePriceCents, &currency)
	if errors.Is(err, pgx.ErrNoRows) {
		writeJSONError(w, http.StatusNotFound, "ruta no encontrada")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "error consultando ruta")
		return
	}

	if req.PickupStopID != nil {
		const pickupQuery = `
			SELECT 1
			FROM app.route_stops
			WHERE id = $1 AND route_id = $2
		`
		var dummy int
		err := pool.QueryRow(ctx, pickupQuery, *req.PickupStopID, req.RouteID).Scan(&dummy)
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSONError(w, http.StatusBadRequest, "pickup_stop_id no pertenece a la ruta")
			return
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "error validando pickup_stop_id")
			return
		}
	}

	if req.DropoffStopID != nil {
		const dropoffQuery = `
			SELECT 1
			FROM app.route_stops
			WHERE id = $1 AND route_id = $2
		`
		var dummy int
		err := pool.QueryRow(ctx, dropoffQuery, *req.DropoffStopID, req.RouteID).Scan(&dummy)
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSONError(w, http.StatusBadRequest, "dropoff_stop_id no pertenece a la ruta")
			return
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "error validando dropoff_stop_id")
			return
		}
	}

	tripID := uuid.New()
	passengerID := uuid.MustParse(dummyUserID)
	now := time.Now().UTC()
	scheduledAt := now.Add(24 * time.Hour)

	const insertTrip = `
		INSERT INTO app.trips (
			id, route_id, passenger_id,
			pickup_stop_id, dropoff_stop_id,
			status, payment_method,
			price_cents, currency,
			scheduled_at, created_at, updated_at
		)
		VALUES (
			$1, $2, $3,
			$4, $5,
			'requested', $6,
			$7, $8,
			$9, $10, $10
		)
		RETURNING
			id, route_id, passenger_id,
			pickup_stop_id, dropoff_stop_id,
			status, payment_method,
			price_cents, currency,
			scheduled_at, created_at, updated_at
	`

	var trip Trip
	err = pool.QueryRow(
		ctx,
		insertTrip,
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
		writeJSONError(w, http.StatusInternalServerError, "error creando viaje")
		return
	}

	writeJSON(w, http.StatusOK, trip)
}

// GET /trips/{id}
func getTripByIDHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pool, err := db.GetPool(ctx)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "error de base de datos")
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		writeJSONError(w, http.StatusBadRequest, "id de viaje requerido")
		return
	}

	tripID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "id de viaje inválido")
		return
	}

	const tripQuery = `
		SELECT
			id, route_id, passenger_id,
			pickup_stop_id, dropoff_stop_id,
			status, payment_method,
			price_cents, currency,
			scheduled_at, created_at, updated_at
		FROM app.trips
		WHERE id = $1
	`

	var trip Trip
	err = pool.QueryRow(ctx, tripQuery, tripID).Scan(
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
	if errors.Is(err, pgx.ErrNoRows) {
		writeJSONError(w, http.StatusNotFound, "viaje no encontrado")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "error consultando viaje")
		return
	}

	const routeQuery = `
		SELECT id, name, origin_name, destination_name, base_price_cents
		FROM app.routes
		WHERE id = $1
	`

	var routeID uuid.UUID
	var routeName, originName, destName string
	var basePriceCents int
	err = pool.QueryRow(ctx, routeQuery, trip.RouteID).Scan(
		&routeID,
		&routeName,
		&originName,
		&destName,
		&basePriceCents,
	)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "error consultando ruta")
		return
	}

	var pickupInfo *StopInfo
	if trip.PickupStopID != nil {
		const pickupQuery = `SELECT id, name FROM app.route_stops WHERE id = $1`
		var stop StopInfo
		err = pool.QueryRow(ctx, pickupQuery, *trip.PickupStopID).Scan(&stop.ID, &stop.Name)
		if err == nil {
			pickupInfo = &stop
		}
	}

	var dropoffInfo *StopInfo
	if trip.DropoffStopID != nil {
		const dropoffQuery = `SELECT id, name FROM app.route_stops WHERE id = $1`
		var stop StopInfo
		err = pool.QueryRow(ctx, dropoffQuery, *trip.DropoffStopID).Scan(&stop.ID, &stop.Name)
		if err == nil {
			dropoffInfo = &stop
		}
	}

	price := float64(trip.PriceCents) / 100.0
	resp := TripDetail{
		ID:          trip.ID,
		PassengerID: trip.PassengerID,
		Route: RouteInfo{
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

	writeJSON(w, http.StatusOK, resp)
}