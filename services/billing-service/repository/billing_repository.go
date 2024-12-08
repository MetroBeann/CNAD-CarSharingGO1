package repository

import (
    "database/sql"
    "fmt"
    "time"
    
    "billing-service/models"
)


// Updated Invoice struct to match database schema
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
    VehicleModel   string    `json:"vehicle_model"`
}

type BillingRepository struct {
    db *sql.DB
}

// User struct for authentication
type User struct {
    ID            int
    Email         string
    PasswordHash  string
    PhoneNumber   string
    MembershipTier string
}

func NewBillingRepository(db *sql.DB) *BillingRepository {
    return &BillingRepository{db: db}
}

// User-related methods
func (r *BillingRepository) GetUserByEmail(email string) (*User, error) {
    var user User
    err := r.db.QueryRow(`
        SELECT id, email, password_hash, phone_number, membership_tier 
        FROM users 
        WHERE email = $1`,
        email,
    ).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.PhoneNumber, &user.MembershipTier)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *BillingRepository) CreateUser(email, passwordHash, phoneNumber, membershipTier string) (*User, error) {
    var user User
    err := r.db.QueryRow(`
        INSERT INTO users (email, password_hash, phone_number, membership_tier)
        VALUES ($1, $2, $3, $4)
        RETURNING id, email, password_hash, phone_number, membership_tier`,
        email, passwordHash, phoneNumber, membershipTier,
    ).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.PhoneNumber, &user.MembershipTier)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// Billing-related methods
func (r *BillingRepository) CalculateRentalCost(userID int, duration float64) (*models.BillingCalculation, error) {
    var membershipTier string
    var hourlyRate, discount float64

    // Get user's membership tier
    err := r.db.QueryRow(`
        SELECT membership_tier 
        FROM users 
        WHERE id = $1
    `, userID).Scan(&membershipTier)
    if err != nil {
        return nil, fmt.Errorf("error getting user membership: %v", err)
    }

    // Get pricing for the user's tier
    err = r.db.QueryRow(`
        SELECT hourly_rate, discount
        FROM pricing_tiers
        WHERE name = $1
    `, membershipTier).Scan(&hourlyRate, &discount)
    if err != nil {
        return nil, fmt.Errorf("error getting pricing info: %v", err)
    }

    baseRate := hourlyRate * duration
    memberDiscount := discount * duration // Flat discount per hour
    finalAmount := baseRate - memberDiscount

    return &models.BillingCalculation{
        Duration:       duration,
        BaseRate:      baseRate,
        MemberDiscount: memberDiscount,
        FinalAmount:   finalAmount,
    }, nil
}

func (r *BillingRepository) CreateInvoice(userID, bookingID int, duration float64) (*models.Invoice, error) {
    // Calculate costs
    calculation, err := r.CalculateRentalCost(userID, duration)
    if err != nil {
        return nil, err
    }

    // Begin transaction
    tx, err := r.db.Begin()
    if err != nil {
        return nil, err
    }

    var invoice models.Invoice
    err = tx.QueryRow(`
        INSERT INTO invoices (
            user_id, booking_id, amount, discount_amount, 
            final_amount, status, created_at
        )
        VALUES ($1, $2, $3, $4, $5, 'pending', CURRENT_TIMESTAMP)
        RETURNING id, user_id, booking_id, amount, discount_amount, 
                  final_amount, status, created_at
    `, userID, bookingID, calculation.BaseRate, calculation.MemberDiscount, 
       calculation.FinalAmount).Scan(
        &invoice.ID, &invoice.UserID, &invoice.BookingID, &invoice.Amount,
        &invoice.DiscountAmount, &invoice.FinalAmount, &invoice.Status, &invoice.CreatedAt)
    
    if err != nil {
        tx.Rollback()
        return nil, fmt.Errorf("error creating invoice: %v", err)
    }

    // Update booking with final cost
    _, err = tx.Exec(`
        UPDATE bookings 
        SET total_cost = $1 
        WHERE id = $2
    `, calculation.FinalAmount, bookingID)

    if err != nil {
        tx.Rollback()
        return nil, fmt.Errorf("error updating booking cost: %v", err)
    }

    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("error committing transaction: %v", err)
    }

    return &invoice, nil
}

// retrieves all invoices for a specific user
func (r *BillingRepository) GetUserInvoices(userID int) ([]Invoice, error) {
    query := `
        SELECT 
            i.id,
            i.booking_id,
            i.user_id,
            i.amount,
            i.discount_amount,
            i.final_amount,
            i.status,
            i.created_at,
            i.updated_at,
            b.vehicle_model  -- Assuming you have this in your bookings table
        FROM invoices i
        LEFT JOIN bookings b ON i.booking_id = b.id
        WHERE i.user_id = $1
        ORDER BY i.created_at DESC
    `
    
    rows, err := r.db.Query(query, userID)
    if err != nil {
        return nil, fmt.Errorf("error querying invoices: %v", err)
    }
    defer rows.Close()

    var invoices []Invoice
    for rows.Next() {
        var inv Invoice
        err := rows.Scan(
            &inv.ID,
            &inv.BookingID,
            &inv.UserID,
            &inv.Amount,
            &inv.DiscountAmount,
            &inv.FinalAmount,
            &inv.Status,
            &inv.CreatedAt,
            &inv.UpdatedAt,
            &inv.VehicleModel,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning invoice row: %v", err)
        }
        invoices = append(invoices, inv)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating invoice rows: %v", err)
    }

    return invoices, nil
}


func (r *BillingRepository) AddPaymentMethod(pm *models.PaymentMethod) error {
    // Begin transaction
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }

    // unset any existing default
    if pm.IsDefault {
        _, err = tx.Exec(`
            UPDATE payment_methods 
            SET is_default = false 
            WHERE user_id = $1 AND is_default = true
        `, pm.UserID)
        if err != nil {
            tx.Rollback()
            return err
        }
    }

    // Insert new payment method
    err = tx.QueryRow(`
        INSERT INTO payment_methods (
            user_id, type, provider, last_four, expiry_date, 
            is_default, created_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
        RETURNING id, created_at
    `, pm.UserID, pm.Type, pm.Provider, pm.LastFour, 
       pm.ExpiryDate, pm.IsDefault).
    Scan(&pm.ID, &pm.CreatedAt)

    if err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
}

// updates the invoice status to paid
func (r *BillingRepository) ProcessPayment(invoiceID int) error {
    query := `
        UPDATE invoices 
        SET 
            status = 'paid',
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1 AND status = 'pending'
        RETURNING id
    `
    
    var id int
    err := r.db.QueryRow(query, invoiceID).Scan(&id)
    if err != nil {
        if err == sql.ErrNoRows {
            return fmt.Errorf("invoice not found or already paid")
        }
        return fmt.Errorf("error processing payment: %v", err)
    }
    
    return nil
}