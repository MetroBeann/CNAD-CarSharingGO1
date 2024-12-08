package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "time"
    "fmt"

    "github.com/gorilla/mux"
    "vehicle-service/models"
    "vehicle-service/repository"
)

type VehicleHandler struct {
    repo *repository.VehicleRepository
}

func NewVehicleHandler(repo *repository.VehicleRepository) *VehicleHandler {
    return &VehicleHandler{repo: repo}
}

// Response wrapper
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

// sendJSON helper
func sendJSON(w http.ResponseWriter, status int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(payload); err != nil {
        log.Printf("Error sending response: %v", err)
    }
}


func (h *VehicleHandler) GetAvailableVehicles(w http.ResponseWriter, r *http.Request) {
    log.Printf("Received request for available vehicles")
    
    startTime := r.URL.Query().Get("start_time")
    endTime := r.URL.Query().Get("end_time")
    
    log.Printf("Query parameters - start_time: %s, end_time: %s", startTime, endTime)
    
    if startTime == "" || endTime == "" {
        log.Printf("Missing required query parameters")
        sendJSON(w, http.StatusBadRequest, Response{
            Success: false,
            Error: "start_time and end_time query parameters are required",
        })
        return
    }

    start, err := time.Parse(time.RFC3339, startTime)
    if err != nil {
        log.Printf("Error parsing start_time: %v", err)
        sendJSON(w, http.StatusBadRequest, Response{
            Success: false,
            Error: fmt.Sprintf("invalid start_time format: %v", err),
        })
        return
    }

    end, err := time.Parse(time.RFC3339, endTime)
    if err != nil {
        log.Printf("Error parsing end_time: %v", err)
        sendJSON(w, http.StatusBadRequest, Response{
            Success: false,
            Error: fmt.Sprintf("invalid end_time format: %v", err),
        })
        return
    }

    vehicles, err := h.repo.GetAvailableVehicles(start, end)
    if err != nil {
        log.Printf("Error getting available vehicles: %v", err)
        sendJSON(w, http.StatusInternalServerError, Response{
            Success: false,
            Error: fmt.Sprintf("failed to get available vehicles: %v", err),
        })
        return
    }

    log.Printf("Successfully retrieved %d vehicles", len(vehicles))
    sendJSON(w, http.StatusOK, Response{
        Success: true,
        Data: vehicles,
    })
}

func (h *VehicleHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
    var req struct {
        VehicleID int       `json:"vehicle_id"`
        StartTime time.Time `json:"start_time"`
        EndTime   time.Time `json:"end_time"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendJSON(w, http.StatusBadRequest, Response{
            Success: false,
            Error: "invalid request body",
        })
        return
    }

    userID, ok := r.Context().Value("user_id").(int)
    if !ok {
        sendJSON(w, http.StatusUnauthorized, Response{
            Success: false,
            Error: "unauthorized",
        })
        return
    }

    booking := &models.Booking{
        UserID:    userID,
        VehicleID: req.VehicleID,
        StartTime: req.StartTime,
        EndTime:   req.EndTime,
        Status:    "pending",
    }

    if err := h.repo.CreateReservation(booking); err != nil {
        log.Printf("Error creating booking: %v", err)
        sendJSON(w, http.StatusInternalServerError, Response{
            Success: false,
            Error: "failed to create booking",
        })
        return
    }

    sendJSON(w, http.StatusCreated, Response{
        Success: true,
        Data: booking,
    })
}

func (h *VehicleHandler) GetUserBookings(w http.ResponseWriter, r *http.Request) {
    userID, ok := r.Context().Value("user_id").(int)
    if !ok {
        sendJSON(w, http.StatusUnauthorized, Response{
            Success: false,
            Error: "unauthorized",
        })
        return
    }

    bookings, err := h.repo.GetUserReservations(userID)
    if err != nil {
        log.Printf("Error getting user bookings: %v", err)
        sendJSON(w, http.StatusInternalServerError, Response{
            Success: false,
            Error: "failed to get bookings",
        })
        return
    }

    sendJSON(w, http.StatusOK, Response{
        Success: true,
        Data: bookings,
    })
}

func (h *VehicleHandler) UpdateBooking(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    bookingID, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendJSON(w, http.StatusBadRequest, Response{
            Success: false,
            Error: "invalid booking ID",
        })
        return
    }

    userID, ok := r.Context().Value("user_id").(int)
    if !ok {
        sendJSON(w, http.StatusUnauthorized, Response{
            Success: false,
            Error: "unauthorized",
        })
        return
    }

    var req struct {
        StartTime *time.Time `json:"start_time,omitempty"`
        EndTime   *time.Time `json:"end_time,omitempty"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendJSON(w, http.StatusBadRequest, Response{
            Success: false,
            Error: "invalid request body",
        })
        return
    }

    // Verify ownership
    if err := h.repo.VerifyBookingOwnership(bookingID, userID); err != nil {
        sendJSON(w, http.StatusNotFound, Response{
            Success: false,
            Error: "booking not found or unauthorized",
        })
        return
    }

    if err := h.repo.UpdateReservation(bookingID, req.StartTime, req.EndTime); err != nil {
        log.Printf("Error updating booking: %v", err)
        sendJSON(w, http.StatusInternalServerError, Response{
            Success: false,
            Error: "failed to update booking",
        })
        return
    }

    sendJSON(w, http.StatusOK, Response{
        Success: true,
        Data: map[string]string{"message": "booking updated successfully"},
    })
}

func (h *VehicleHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    bookingID, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendJSON(w, http.StatusBadRequest, Response{
            Success: false,
            Error: "invalid booking ID",
        })
        return
    }

    userID, ok := r.Context().Value("user_id").(int)
    if !ok {
        sendJSON(w, http.StatusUnauthorized, Response{
            Success: false,
            Error: "unauthorized",
        })
        return
    }

    // Verify ownership
    if err := h.repo.VerifyBookingOwnership(bookingID, userID); err != nil {
        sendJSON(w, http.StatusNotFound, Response{
            Success: false,
            Error: "booking not found or unauthorized",
        })
        return
    }

    if err := h.repo.UpdateBookingStatus(bookingID, "cancelled"); err != nil {
        log.Printf("Error cancelling booking: %v", err)
        sendJSON(w, http.StatusInternalServerError, Response{
            Success: false,
            Error: "failed to cancel booking",
        })
        return
    }

    sendJSON(w, http.StatusOK, Response{
        Success: true,
        Data: map[string]string{"message": "booking cancelled successfully"},
    })
}

func (h *VehicleHandler) UpdateVehicleStatus(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    vehicleID, err := strconv.Atoi(vars["id"])
    if err != nil {
        sendJSON(w, http.StatusBadRequest, Response{
            Success: false,
            Error: "invalid vehicle ID",
        })
        return
    }

    var update struct {
        Location         *string `json:"location,omitempty"`
        BatteryLevel    *int    `json:"battery_level,omitempty"`
        CleanlinessStatus *string `json:"cleanliness_status,omitempty"`
    }

    if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
        sendJSON(w, http.StatusBadRequest, Response{
            Success: false,
            Error: "invalid request body",
        })
        return
    }

    if err := h.repo.UpdateVehicleStatus(vehicleID, update.Location, update.BatteryLevel, update.CleanlinessStatus); err != nil {
        log.Printf("Error updating vehicle status: %v", err)
        sendJSON(w, http.StatusInternalServerError, Response{
            Success: false,
            Error: "failed to update vehicle status",
        })
        return
    }

    sendJSON(w, http.StatusOK, Response{
        Success: true,
        Data: map[string]string{"message": "vehicle status updated successfully"},
    })
}