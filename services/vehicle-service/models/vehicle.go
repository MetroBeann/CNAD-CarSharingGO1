package models

import "time"

type Vehicle struct {
    ID               int       `json:"id"`
    Model           string    `json:"model"`
    Type            string    `json:"type"`
    LicensePlate    string    `json:"license_plate"`
    Status          string    `json:"status"`
    Location        *string   `json:"location"`
    BatteryLevel    *int      `json:"battery_level"`
    CleanlinessStatus *string `json:"cleanliness_status"`
    CreatedAt       time.Time `json:"created_at"`
    LastStatusUpdate time.Time `json:"last_status_update"`
}

type Booking struct {
    ID          int       `json:"id"`
    UserID      int       `json:"user_id"`
    VehicleID   int       `json:"vehicle_id"`
    VehicleModel string   `json:"vehicle_model,omitempty"` // Added for frontend display
    StartTime   time.Time `json:"start_time"`
    EndTime     time.Time `json:"end_time"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
