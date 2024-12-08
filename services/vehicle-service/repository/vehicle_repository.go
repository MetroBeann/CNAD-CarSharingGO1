package repository

import (
    "database/sql"
    "errors"
    "time"
    "vehicle-service/models"
    "fmt"
)

type VehicleRepository struct {
    db *sql.DB
}

func NewVehicleRepository(db *sql.DB) *VehicleRepository {
    return &VehicleRepository{db: db}
}

func (r *VehicleRepository) GetAvailableVehicles(startTime, endTime time.Time) ([]models.Vehicle, error) {
    query := `
        SELECT id, model, type, license_plate, status, location, 
               battery_level, cleanliness_status, created_at, last_status_update
        FROM vehicles v
        WHERE v.id NOT IN (
            SELECT vehicle_id 
            FROM bookings 
            WHERE status IN ('pending', 'confirmed')
            AND (
                (start_time <= $1 AND end_time >= $1)
                OR (start_time <= $2 AND end_time >= $2)
                OR (start_time >= $1 AND end_time <= $2)
            )
        )
        AND v.status = 'available'
        AND (v.battery_level IS NULL OR v.battery_level >= 20)
        AND (v.cleanliness_status IS NULL OR v.cleanliness_status != 'needs_cleaning')
    `
    
    rows, err := r.db.Query(query, startTime, endTime)
    if err != nil {
        return nil, fmt.Errorf("error querying available vehicles: %v", err)
    }
    defer rows.Close()

    var vehicles []models.Vehicle
    for rows.Next() {
        var v models.Vehicle
        err := rows.Scan(
            &v.ID, &v.Model, &v.Type, &v.LicensePlate, &v.Status,
            &v.Location, &v.BatteryLevel, &v.CleanlinessStatus,
            &v.CreatedAt, &v.LastStatusUpdate,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning vehicle row: %v", err)
        }
        vehicles = append(vehicles, v)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating vehicle rows: %v", err)
    }

    return vehicles, nil
}

func (r *VehicleRepository) GetVehicleByID(vehicleID int) (*models.Vehicle, error) {
    query := `
        SELECT id, model, type, license_plate, status, location, 
               battery_level, cleanliness_status, created_at, last_status_update
        FROM vehicles
        WHERE id = $1
    `
    
    var v models.Vehicle
    err := r.db.QueryRow(query, vehicleID).Scan(
        &v.ID, &v.Model, &v.Type, &v.LicensePlate, &v.Status,
        &v.Location, &v.BatteryLevel, &v.CleanlinessStatus,
        &v.CreatedAt, &v.LastStatusUpdate,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.New("vehicle not found")
        }
        return nil, err
    }
    
    return &v, nil
}

func (r *VehicleRepository) CreateReservation(booking *models.Booking) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }

    query := `
        INSERT INTO bookings (user_id, vehicle_id, start_time, end_time, status)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at
    `
    
    err = tx.QueryRow(
        query,
        booking.UserID,
        booking.VehicleID,
        booking.StartTime,
        booking.EndTime,
        booking.Status,
    ).Scan(&booking.ID, &booking.CreatedAt, &booking.UpdatedAt)

    if err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
}

func (r *VehicleRepository) GetUserReservations(userID int) ([]models.Booking, error) {
    query := `
        SELECT b.id, b.user_id, b.vehicle_id, v.model as vehicle_model,
               b.start_time, b.end_time, b.status, b.created_at, b.updated_at
        FROM bookings b
        JOIN vehicles v ON b.vehicle_id = v.id
        WHERE b.user_id = $1
        ORDER BY b.created_at DESC
    `
    
    rows, err := r.db.Query(query, userID)
    if err != nil {
        return nil, fmt.Errorf("error querying user bookings: %v", err)
    }
    defer rows.Close()

    var bookings []models.Booking
    for rows.Next() {
        var b models.Booking
        err := rows.Scan(
            &b.ID, &b.UserID, &b.VehicleID, &b.VehicleModel,
            &b.StartTime, &b.EndTime, &b.Status,
            &b.CreatedAt, &b.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("error scanning booking row: %v", err)
        }
        bookings = append(bookings, b)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating booking rows: %v", err)
    }

    return bookings, nil
}

func (r *VehicleRepository) UpdateReservation(bookingID int, startTime, endTime *time.Time) error {
    query := `
        UPDATE bookings
        SET start_time = COALESCE($1, start_time),
            end_time = COALESCE($2, end_time),
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $3 AND status IN ('pending', 'confirmed')
    `
    
    result, err := r.db.Exec(query, startTime, endTime, bookingID)
    if err != nil {
        return err
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rows == 0 {
        return errors.New("booking not found or cannot be updated")
    }

    return nil
}

func (r *VehicleRepository) UpdateBookingStatus(bookingID int, status string) error {
    query := `
        UPDATE bookings
        SET status = $1,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $2 AND status IN ('pending', 'confirmed')
    `
    
    result, err := r.db.Exec(query, status, bookingID)
    if err != nil {
        return err
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rows == 0 {
        return errors.New("booking not found or cannot be updated")
    }

    return nil
}

func (r *VehicleRepository) UpdateVehicleStatus(vehicleID int, location *string, batteryLevel *int, cleanlinessStatus *string) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }

    // Update current status
    updateQuery := `
        UPDATE vehicles 
        SET location = COALESCE($1, location),
            battery_level = COALESCE($2, battery_level),
            cleanliness_status = COALESCE($3, cleanliness_status),
            last_status_update = CURRENT_TIMESTAMP
        WHERE id = $4
    `
    
    result, err := tx.Exec(updateQuery, location, batteryLevel, cleanlinessStatus, vehicleID)
    if err != nil {
        tx.Rollback()
        return err
    }

    rows, err := result.RowsAffected()
    if err != nil {
        tx.Rollback()
        return err
    }

    if rows == 0 {
        tx.Rollback()
        return errors.New("vehicle not found")
    }

    // Record history
    historyQuery := `
        INSERT INTO vehicle_status_history (vehicle_id, location, battery_level, cleanliness_status)
        VALUES ($1, $2, $3, $4)
    `
    
    _, err = tx.Exec(historyQuery, vehicleID, location, batteryLevel, cleanlinessStatus)
    if err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
}

func (r *VehicleRepository) VerifyBookingOwnership(bookingID, userID int) error {
    var count int
    err := r.db.QueryRow(`
        SELECT COUNT(*) 
        FROM bookings 
        WHERE id = $1 AND user_id = $2
    `, bookingID, userID).Scan(&count)
    
    if err != nil {
        return err
    }
    
    if count == 0 {
        return errors.New("booking not found or unauthorized")
    }
    
    return nil
}