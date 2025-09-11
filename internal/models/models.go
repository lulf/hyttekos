package models

import (
	"time"
)

// TemperatureReading represents a temperature measurement
type TemperatureReading struct {
	ID          int       `json:"id" db:"id"`
	Temperature float64   `json:"temperature" db:"temperature"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Session represents a user session with OAuth tokens
type Session struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	AccessToken  string    `json:"access_token" db:"access_token"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// GadgetStatus represents the status of a gadget (heating/hot water)
type GadgetStatus struct {
	Enabled     bool      `json:"enabled"`
	LastUpdated time.Time `json:"last_updated"`
}

// SystemStatus represents the overall status of all cabin systems
type SystemStatus struct {
	Heating     GadgetStatus       `json:"heating"`
	HotWater    GadgetStatus       `json:"hot_water"`
	Temperature TemperatureReading `json:"temperature"`
}

// GadgetAction represents an action to perform on a gadget
type GadgetAction struct {
	Action string `json:"action" binding:"required,oneof=on off"`
}

// GadgetActionResponse represents the response after performing a gadget action
type GadgetActionResponse struct {
	Success   bool         `json:"success"`
	Heating   *GadgetStatus `json:"heating,omitempty"`
	HotWater  *GadgetStatus `json:"hot_water,omitempty"`
	Error     string       `json:"error,omitempty"`
	Message   string       `json:"message,omitempty"`
}

// TemperatureSubmission represents temperature data submitted by sensors
type TemperatureSubmission struct {
	Value     float64   `json:"value" binding:"required"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// APIError represents an API error response
type APIError struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}