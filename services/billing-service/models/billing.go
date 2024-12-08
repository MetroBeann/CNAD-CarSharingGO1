package models

import (
    "time"
)

type PricingTier struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`  // Basic, Premium, VIP
    HourlyRate  float64 `json:"hourly_rate"`
    Discount    float64 `json:"discount"` // Percentage discount
    CreatedAt   time.Time `json:"created_at"`
}

type Invoice struct {
    ID             int       `json:"id"`
    BookingID      int       `json:"booking_id"`
    UserID         int       `json:"user_id"`
    Amount         float64   `json:"amount"`
    DiscountAmount float64   `json:"discount_amount"`
    FinalAmount    float64   `json:"final_amount"`
    Status         string    `json:"status"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
    VehicleID      int       `json:"vehicle_id"`
    VehicleModel   string    `json:"vehicle_model"`
    MembershipTier string    `json:"membership_tier"`
}

type PaymentMethod struct {
    ID          int       `json:"id"`
    UserID      int       `json:"user_id"`
    Type        string    `json:"type"`     // credit_card, debit_card, etc.
    Provider    string    `json:"provider"` // visa, mastercard, etc.
    LastFour    string    `json:"last_four"`
    ExpiryDate  string    `json:"expiry_date"`
    IsDefault   bool      `json:"is_default"`
    CreatedAt   time.Time `json:"created_at"`
}

type BillingCalculation struct {
    Duration        float64 `json:"duration"`        // in hours
    BaseRate        float64 `json:"base_rate"`
    MemberDiscount  float64 `json:"member_discount"`
    FinalAmount     float64 `json:"final_amount"`
}

// Response wrapper for consistent API responses
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string     `json:"error,omitempty"`
}