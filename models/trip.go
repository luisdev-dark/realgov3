package models

import (
	"time"

	"github.com/google/uuid"
)

type Trip struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	RouteID       uuid.UUID  `json:"route_id" db:"route_id"`
	PassengerID   uuid.UUID  `json:"passenger_id" db:"passenger_id"`
	PickupStopID  *uuid.UUID `json:"pickup_stop_id" db:"pickup_stop_id"`
	DropoffStopID *uuid.UUID `json:"dropoff_stop_id" db:"dropoff_stop_id"`
	Status        string     `json:"status" db:"status"` // requested, confirmed, completed, cancelled
	PaymentMethod string     `json:"payment_method" db:"payment_method"` // cash, yape, pling
	PriceCents    int        `json:"price_cents" db:"price_cents"`
	Currency      string     `json:"currency" db:"currency"`
	ScheduledAt   *time.Time `json:"scheduled_at" db:"scheduled_at"`
	StartedAt     *time.Time `json:"started_at" db:"started_at"`
	FinishedAt    *time.Time `json:"finished_at" db:"finished_at"`
	CancelledAt   *time.Time `json:"cancelled_at" db:"cancelled_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// TripDetail es la respuesta completa de GET /trips/{id}
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

type RouteInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Origin      string    `json:"origin"`
	Destination string    `json:"destination"`
	BasePrice   float64   `json:"base_price"`
}
