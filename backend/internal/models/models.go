package models

import "time"

const (
	CategoryDelivery   = "delivery_zone"
	CategoryRestricted = "restricted_zone"
	CategoryToll       = "toll_zone"
	CategoryCustomer   = "customer_area"

	EventEntry = "entry"
	EventExit  = "exit"
	EventBoth  = "both"
)

func ValidCategory(c string) bool {
	switch c {
	case CategoryDelivery, CategoryRestricted, CategoryToll, CategoryCustomer:
		return true
	}
	return false
}

func ValidEventType(e string) bool {
	switch e {
	case EventEntry, EventExit, EventBoth:
		return true
	}
	return false
}

// Geofence is the row representation. Coordinates are stored separately as
// the original lat/lng pairs the client sent (PostGIS holds them in `geom`
// in lng/lat order for query purposes).
type Geofence struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Coordinates [][]float64 `json:"coordinates"`
	Category    string      `json:"category"`
	Status      string      `json:"status,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

type Vehicle struct {
	ID            string    `json:"id"`
	VehicleNumber string    `json:"vehicle_number"`
	DriverName    string    `json:"driver_name"`
	VehicleType   string    `json:"vehicle_type"`
	Phone         string    `json:"phone"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type AlertConfig struct {
	AlertID       string    `json:"alert_id"`
	GeofenceID    string    `json:"geofence_id"`
	GeofenceName  string    `json:"geofence_name,omitempty"`
	VehicleID     *string   `json:"vehicle_id,omitempty"`
	VehicleNumber *string   `json:"vehicle_number,omitempty"`
	EventType     string    `json:"event_type"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type Violation struct {
	ID            string    `json:"id"`
	VehicleID     string    `json:"vehicle_id"`
	VehicleNumber string    `json:"vehicle_number"`
	GeofenceID    string    `json:"geofence_id"`
	GeofenceName  string    `json:"geofence_name"`
	EventType     string    `json:"event_type"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	Timestamp     time.Time `json:"timestamp"`
}

// AlertEnvelope is the message format pushed over the WebSocket.
type AlertEnvelope struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	Vehicle   struct {
		VehicleID     string `json:"vehicle_id"`
		VehicleNumber string `json:"vehicle_number"`
		DriverName    string `json:"driver_name"`
	} `json:"vehicle"`
	Geofence struct {
		GeofenceID   string `json:"geofence_id"`
		GeofenceName string `json:"geofence_name"`
		Category     string `json:"category"`
	} `json:"geofence"`
	Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
}
