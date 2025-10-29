package models

import "time"

// Device representa un dispositivo GPS en Traccar
type Device struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	UniqueID    string                 `json:"uniqueId"`
	Status      string                 `json:"status"`
	Disabled    bool                   `json:"disabled"`
	LastUpdate  *time.Time             `json:"lastUpdate"`
	PositionID  *int                   `json:"positionId"`
	GroupID     *int                   `json:"groupId"`
	Phone       string                 `json:"phone"`
	Model       string                 `json:"model"`
	Contact     string                 `json:"contact"`
	Category    string                 `json:"category"`
	Attributes  map[string]interface{} `json:"attributes"`
}

// DeviceRequest para crear/actualizar dispositivos
type DeviceRequest struct {
	Name       string                 `json:"name" binding:"required"`
	UniqueID   string                 `json:"uniqueId" binding:"required"`
	Phone      string                 `json:"phone"`
	Model      string                 `json:"model"`
	Contact    string                 `json:"contact"`
	Category   string                 `json:"category"`
	GroupID    *int                   `json:"groupId"`
	Attributes map[string]interface{} `json:"attributes"`
}

// DeviceAccumulators para actualizar distancia y horas
type DeviceAccumulators struct {
	DeviceID      int     `json:"deviceId"`
	TotalDistance float64 `json:"totalDistance"` // en metros
	Hours         float64 `json:"hours"`
}