package models

import (
	"time"

	"github.com/google/uuid"
)

type RouteStop struct {
	ID        uuid.UUID `json:"id" db:"id"`
	RouteID   uuid.UUID `json:"route_id" db:"route_id"`
	Name      string    `json:"name" db:"name"`
	Order     int       `json:"order" db:"order"`
	Latitude  float64   `json:"latitude" db:"latitude"`
	Longitude float64   `json:"longitude" db:"longitude"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
