package models

import "time"

type User struct {
    ID             int       `json:"id"`
    Email          string    `json:"email"`
    PhoneNumber    string    `json:"phone_number"`
    PasswordHash   string    `json:"-"` // Hide from JSON responses
    MembershipTier string    `json:"membership_tier"`
    CreatedAt      time.Time `json:"created_at"`
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type UpdateProfileRequest struct {
    Email          string `json:"email"`
    PhoneNumber    string `json:"phone_number"`
    MembershipTier string `json:"membership_tier"`
}