package models

import (
	"net/http"
	"time"
)

type User struct {
	ID             int                    `json:"id"`
	Name           string                 `json:"name"`
	Login          string                 `json:"login"`
	Email          string                 `json:"email"`
	Phone          string                 `json:"phone"`
	Readonly       bool                   `json:"readonly"`
	Administrator  bool                   `json:"administrator"`
	Map            string                 `json:"map"`
	Latitude       float64                `json:"latitude"`
	Longitude      float64                `json:"longitude"`
	Zoom           int                    `json:"zoom"`
	CoordinateFormat string              `json:"coordinateFormat"`
	Disabled       bool                   `json:"disabled"`
	ExpirationTime *time.Time             `json:"expirationTime"`
	DeviceLimit    int                    `json:"deviceLimit"`
	UserLimit      int                    `json:"userLimit"`
	DeviceReadonly bool                   `json:"deviceReadonly"`
	LimitCommands  bool                   `json:"limitCommands"`
	DisableReports bool                   `json:"disableReports"`
	FixedEmail     bool                   `json:"fixedEmail"`
	POILayer       string                 `json:"poiLayer"`
	TOTPKey        string                 `json:"totpKey"`
	Temporary      bool                   `json:"temporary"`
	Password       string                 `json:"password"`
	Attributes     map[string]interface{} `json:"attributes"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User    *User           `json:"user"`
	Token   string          `json:"token"`
	Cookies []*http.Cookie  `json:"-"` // No incluir en JSON
}