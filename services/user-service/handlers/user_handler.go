package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "time"
    "fmt"

    "github.com/golang-jwt/jwt"
    "golang.org/x/crypto/bcrypt"
    "github.com/gorilla/mux"
    "cnad-carsharinggo/services/user-service/models"
    "cnad-carsharinggo/services/user-service/repository"
)

var jwtKey = []byte("your-secret-key") 

type UserHandler struct {
    UserRepo *repository.UserRepository
}

func NewUserHandler(repo *repository.UserRepository) *UserHandler {
    return &UserHandler{UserRepo: repo}
}

type RegisterRequest struct {
    Email          string `json:"email"`
    Password       string `json:"password"`
    PhoneNumber    string `json:"phone_number"`
    MembershipTier string `json:"membership_tier"`
}

type Claims struct {
    UserID int    `json:"user_id"`
    Email  string `json:"email"`
    jwt.StandardClaims
}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
    var regRequest RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&regRequest); err != nil {
        log.Printf("Decode error: %v", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if regRequest.Email == "" || regRequest.Password == "" {
        http.Error(w, "Email and password are required", http.StatusBadRequest)
        return
    }

    user := &models.User{
        Email:          regRequest.Email,
        PhoneNumber:    regRequest.PhoneNumber,
        MembershipTier: regRequest.MembershipTier,
    }

    if err := h.UserRepo.CreateUser(user, regRequest.Password); err != nil {
        log.Printf("CreateUser error: %v", err)
        http.Error(w, "Failed to register user: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "User registered successfully",
        "user_id": user.ID,
        "email":   user.Email,
    })
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
    var loginRequest models.LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    user, storedPassword, err := h.UserRepo.FindByEmail(loginRequest.Email)
    if err != nil {
        http.Error(w, "User not found", http.StatusUnauthorized)
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(loginRequest.Password)); err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Create claims for JWT
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        UserID: user.ID,
        Email:  user.Email,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    // Generate JWT token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        http.Error(w, "Could not generate token", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "user_id":         user.ID,
        "email":           user.Email,
        "membership_tier": user.MembershipTier,
        "token":          tokenString,
    })
}

func (h *UserHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
    // Get user ID from path parameters
    vars := mux.Vars(r)
    userID := vars["id"]

    // Get claims from context (set by middleware)
    claims, ok := r.Context().Value("claims").(*Claims)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Convert path userID to int for comparison
    requestedUserID := fmt.Sprintf("%d", claims.UserID)
    if userID != requestedUserID {
        http.Error(w, "Unauthorized to update this profile", http.StatusForbidden)
        return
    }

    var updateRequest models.UpdateProfileRequest
    if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if err := h.UserRepo.UpdateProfile(userID, updateRequest); err != nil {
        http.Error(w, "Failed to update profile: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Profile updated successfully",
    })
}

// helper function that can be used by middleware
func ValidateToken(tokenString string) (*Claims, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })

    if err != nil {
        return nil, err
    }

    if !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }

    return claims, nil
}