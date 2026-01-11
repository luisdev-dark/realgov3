package models

import (
	"time"

	"github.com/google/uuid"
)

type Route struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	OriginName      string    `json:"origin_name" db:"origin_name"`
	OriginLat       float64   `json:"origin_lat" db:"origin_lat"`
	OriginLon       float64   `json:"origin_lon" db:"origin_lon"`
	DestinationName string    `json:"destination_name" db:"destination_name"`
	DestinationLat  float64   `json:"destination_lat" db:"destination_lat"`
	DestinationLon  float64   `json:"destination_lon" db:"destination_lon"`
	BasePriceCents  int       `json:"base_price_cents" db:"base_price_cents"`
	Currency        string    `json:"currency" db:"currency"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// RouteDetail es la respuesta completa de GET /routes/{id}
type RouteDetail struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Origin      string     `json:"origin"`
	Destination string     `json:"destination"`
	BasePrice   float64    `json:"base_price"`
	Stops       []StopInfo `json:"stops"`
}

type StopInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
