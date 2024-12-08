package handlers

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "time"
    
    "github.com/gorilla/mux"
    "golang.org/x/crypto/bcrypt"
    jwtgo "github.com/dgrijalva/jwt-go"
    
    "billing-service/models"
    "billing-service/repository"
)

type BillingHandler struct {
    repo *repository.BillingRepository
    jwtSecret []byte
}

func NewBillingHandler(repo *repository.BillingRepository) *BillingHandler {
    return &BillingHandler{
        repo: repo,
        jwtSecret: []byte("your-secret-key"), 
    }
}

// Auth types
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Token         string `json:"token"`
    UserID        int    `json:"user_id"`
    Email         string `json:"email"`
    MembershipTier string `json:"membership_tier"`
}

//  handles user authentication
func (h *BillingHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    user, err := h.repo.GetUserByEmail(req.Email)
    if err != nil {
        sendError(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        sendError(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, jwtgo.MapClaims{
        "user_id": user.ID,
        "email": user.Email,
        "membership_tier": user.MembershipTier,
        "exp": time.Now().Add(24 * time.Hour).Unix(),
    })

    tokenString, err := token.SignedString(h.jwtSecret)
    if err != nil {
        sendError(w, "Error generating token", http.StatusInternalServerError)
        return
    }

    sendJSON(w, http.StatusOK, LoginResponse{
        Token: tokenString,
        UserID: user.ID,
        Email: user.Email,
        MembershipTier: user.MembershipTier,
    })
}


// validates the JWT token
func (h *BillingHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        sendError(w, "No authorization header", http.StatusUnauthorized)
        return
    }

    tokenString := authHeader[7:]

    token, err := jwtgo.Parse(tokenString, func(token *jwtgo.Token) (interface{}, error) {
        return h.jwtSecret, nil
    })

    if err != nil || !token.Valid {
        sendError(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    claims := token.Claims.(jwtgo.MapClaims)
    sendJSON(w, http.StatusOK, map[string]interface{}{
        "valid": true,
        "user_id": claims["user_id"],
        "email": claims["email"],
        "membership_tier": claims["membership_tier"],
    })
}

// handles user registration[]
func (h *BillingHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email         string `json:"email"`
        Password      string `json:"password"`
        PhoneNumber   string `json:"phone_number"`
        MembershipTier string `json:"membership_tier"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        sendError(w, "Error processing password", http.StatusInternalServerError)
        return
    }

    // Create user
    user, err := h.repo.CreateUser(req.Email, string(hashedPassword), req.PhoneNumber, req.MembershipTier)
    if err != nil {
        sendError(w, fmt.Sprintf("Error creating user: %v", err), http.StatusInternalServerError)
        return
    }

    sendJSON(w, http.StatusCreated, map[string]interface{}{
        "success": true,
        "user_id": user.ID,
        "email": user.Email,
    })
}

// calculates estimated cost for a booking
func (h *BillingHandler) CalculateEstimate(w http.ResponseWriter, r *http.Request) {
    var req struct {
        UserID    int     `json:"user_id"`
        StartTime string  `json:"start_time"`
        EndTime   string  `json:"end_time"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Parse times and calculate duration
    start, err := time.Parse(time.RFC3339, req.StartTime)
    if err != nil {
        sendError(w, "Invalid start time format", http.StatusBadRequest)
        return
    }

    end, err := time.Parse(time.RFC3339, req.EndTime)
    if err != nil {
        sendError(w, "Invalid end time format", http.StatusBadRequest)
        return
    }

    duration := end.Sub(start).Hours()
    calculation, err := h.repo.CalculateRentalCost(req.UserID, duration)
    if err != nil {
        sendError(w, fmt.Sprintf("Error calculating cost: %v", err), http.StatusInternalServerError)
        return
    }

    sendJSON(w, http.StatusOK, models.Response{
        Success: true,
        Data: map[string]interface{}{
            "calculation": calculation,
            "breakdown": map[string]interface{}{
                "duration_hours": duration,
                "base_rate_per_hour": calculation.BaseRate / duration,
                "discount_per_hour": calculation.MemberDiscount / duration,
                "total_discount": calculation.MemberDiscount,
                "final_amount": calculation.FinalAmount,
            },
        },
    })
}

// generates a new invoice for a completed booking
func (h *BillingHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
    var req struct {
        UserID    int     `json:"user_id"`
        BookingID int     `json:"booking_id"`
        Duration  float64 `json:"duration"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    invoice, err := h.repo.CreateInvoice(req.UserID, req.BookingID, req.Duration)
    if err != nil {
        sendError(w, fmt.Sprintf("Error creating invoice: %v", err), http.StatusInternalServerError)
        return
    }

    sendJSON(w, http.StatusCreated, models.Response{
        Success: true,
        Data:    invoice,
    })
}

// handles the GET request for user invoices
func (h *BillingHandler) GetUserInvoices(w http.ResponseWriter, r *http.Request) {
    // Extract user ID from URL parameters
    vars := mux.Vars(r)
    userID, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendError(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    // Get invoices from repository
    invoices, err := h.repo.GetUserInvoices(userID)
    if err != nil {
        sendError(w, fmt.Sprintf("Error retrieving invoices: %v", err), http.StatusInternalServerError)
        return
    }

    // Send response
    sendJSON(w, http.StatusOK, models.Response{
        Success: true,
        Data:    invoices,
    })
}

// adds a new payment method for a user
func (h *BillingHandler) AddPaymentMethod(w http.ResponseWriter, r *http.Request) {
    var paymentMethod models.PaymentMethod
    if err := json.NewDecoder(r.Body).Decode(&paymentMethod); err != nil {
        sendError(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if err := h.repo.AddPaymentMethod(&paymentMethod); err != nil {
        sendError(w, fmt.Sprintf("Error adding payment method: %v", err), http.StatusInternalServerError)
        return
    }

    sendJSON(w, http.StatusCreated, models.Response{
        Success: true,
        Data:    paymentMethod,
    })
}

// handles the payment processing for an invoice
func (h *BillingHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
    // Extract invoice ID from URL parameters
    vars := mux.Vars(r)
    invoiceID, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendError(w, "Invalid invoice ID", http.StatusBadRequest)
        return
    }

    // Process the payment
    err = h.repo.ProcessPayment(invoiceID)
    if err != nil {
        sendError(w, fmt.Sprintf("Error processing payment: %v", err), http.StatusInternalServerError)
        return
    }

    // Send success response
    sendJSON(w, http.StatusOK, models.Response{
        Success: true,
        Data: map[string]string{
            "message": "Payment processed successfully",
        },
    })
}

// Helper functions
func sendJSON(w http.ResponseWriter, status int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(payload); err != nil {
        log.Printf("Error encoding response: %v", err)
    }
}

func sendError(w http.ResponseWriter, message string, status int) {
    sendJSON(w, status, models.Response{
        Success: false,
        Error:   message,
    })
}