package repository

import (
    "database/sql"
    "errors"
    "time"
    "fmt"
    "strings"

    "cnad-carsharinggo/services/user-service/models"
    "golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
    DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{DB: db}
}

func (r *UserRepository) CreateUser(user *models.User, password string) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    if user.MembershipTier == "" {
        user.MembershipTier = "Basic"
    }

    query := `
        INSERT INTO users (email, phone_number, password_hash, membership_tier, created_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
    err = r.DB.QueryRow(
        query, 
        user.Email, 
        user.PhoneNumber, 
        string(hashedPassword), 
        user.MembershipTier, 
        time.Now(),
    ).Scan(&user.ID)

    return err
}

func (r *UserRepository) FindByEmail(email string) (*models.User, string, error) {
    var user models.User
    var storedPassword string
    query := `SELECT id, email, password_hash, membership_tier FROM users WHERE email = $1`
    
    err := r.DB.QueryRow(query, email).Scan(
        &user.ID, 
        &user.Email, 
        &storedPassword, 
        &user.MembershipTier,
    )
    if err != nil {
        return nil, "", err
    }

    return &user, storedPassword, nil
}

func (r *UserRepository) UpdateProfile(userID string, updates models.UpdateProfileRequest) error {
    var setClause []string
    var updateValues []interface{}
    paramCount := 1

    // Build dynamic update query based on provided fields
    if updates.Email != "" {
        setClause = append(setClause, fmt.Sprintf("email = $%d", paramCount))
        updateValues = append(updateValues, updates.Email)
        paramCount++
    }

    if updates.PhoneNumber != "" {
        setClause = append(setClause, fmt.Sprintf("phone_number = $%d", paramCount))
        updateValues = append(updateValues, updates.PhoneNumber)
        paramCount++
    }

    if updates.MembershipTier != "" {
        setClause = append(setClause, fmt.Sprintf("membership_tier = $%d", paramCount))
        updateValues = append(updateValues, updates.MembershipTier)
        paramCount++
    }

    // If no updates provided
    if len(setClause) == 0 {
        return errors.New("no updates provided")
    }

    // Construct the final query
    query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d",
        strings.Join(setClause, ", "),
        paramCount)
    
    // Add the userID as the last parameter
    updateValues = append(updateValues, userID)

    // Execute the update
    result, err := r.DB.Exec(query, updateValues...)
    if err != nil {
        return err
    }

    // Check if any rows were affected
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if rowsAffected == 0 {
        return errors.New("user not found")
    }

    return nil
}