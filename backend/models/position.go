package models

import "time"

// Position representa una posición GPS de un dispositivo en Traccar
type Position struct {
	ID          int                    `json:"id"`
	DeviceID    int                    `json:"deviceId"`
	Protocol    string                 `json:"protocol"`
	DeviceTime  *time.Time             `json:"deviceTime"`
	FixTime     *time.Time             `json:"fixTime"`
	ServerTime  *time.Time             `json:"serverTime"`
	Valid       bool                   `json:"valid"`
	Latitude    float64                `json:"latitude"`
	Longitude   float64                `json:"longitude"`
	Altitude    float64                `json:"altitude"`
	Speed       float64                `json:"speed"`       // en nudos
	Course      float64                `json:"course"`      // dirección en grados
	Address     string                 `json:"address"`
	Accuracy    float64                `json:"accuracy"`
	Network     map[string]interface{} `json:"network"`
	GeofenceIDs []int                  `json:"geofenceIds"`
	Attributes  map[string]interface{} `json:"attributes"`
}

// PositionWithDevice combina posición con información del dispositivo para el mapa
type PositionWithDevice struct {
	Position
	DeviceName   string `json:"deviceName"`
	DeviceStatus string `json:"deviceStatus"`
	DeviceModel  string `json:"deviceModel"`
}

// MapData estructura para enviar datos al mapa
type MapData struct {
	Positions []PositionWithDevice `json:"positions"`
	Center    MapCenter            `json:"center"`
	Zoom      int                  `json:"zoom"`
}

// MapCenter para centrar el mapa
type MapCenter struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}