// Path: services/vehicle-service/main.go
package main

import (
    "database/sql"
    "log"
    "net/http"
    "path/filepath"
    "time"
    "fmt"

    _ "github.com/lib/pq"
    gorillaCORS "github.com/gorilla/handlers"
    "github.com/gorilla/mux"
    
    "vehicle-service/handlers"
    "vehicle-service/middleware"
    "vehicle-service/repository"
)

func setupDB() (*sql.DB, error) {
    connStr := "postgres://postgres.wjdhhzmaclmsvaiszagk:22KC6282t04@@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres?sslmode=require"
    
    log.Printf("Attempting to connect to database with connection string: %v", connStr)
    
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("error opening database: %v", err)
    }
    
    if err = db.Ping(); err != nil {
        return nil, fmt.Errorf("error pinging database: %v", err)
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    log.Printf("Successfully connected to database")
    return db, nil
}

func checkDatabaseConnection(db *sql.DB) {
    log.Printf("Testing database connection...")

    // Test vehicles table
    var vehicleCount int
    err := db.QueryRow("SELECT COUNT(*) FROM vehicles").Scan(&vehicleCount)
    if err != nil {
        log.Printf("ERROR: Failed to query vehicles table: %v", err)
    } else {
        log.Printf("SUCCESS: Found %d vehicles in the database", vehicleCount)
    }

    // Test bookings table
    var bookingCount int
    err = db.QueryRow("SELECT COUNT(*) FROM bookings").Scan(&bookingCount)
    if err != nil {
        log.Printf("ERROR: Failed to query bookings table: %v", err)
    } else {
        log.Printf("SUCCESS: Found %d bookings in the database", bookingCount)
    }

    // Test vehicle_status_history table
    var historyCount int
    err = db.QueryRow("SELECT COUNT(*) FROM vehicle_status_history").Scan(&historyCount)
    if err != nil {
        log.Printf("ERROR: Failed to query vehicle_status_history table: %v", err)
    } else {
        log.Printf("SUCCESS: Found %d status history records in the database", historyCount)
    }
}

func setupCORS(handler http.Handler) http.Handler {
    return gorillaCORS.CORS(
        gorillaCORS.AllowedOrigins([]string{"*"}),
        gorillaCORS.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
        gorillaCORS.AllowedHeaders([]string{"Content-Type", "Authorization"}),
        gorillaCORS.MaxAge(86400),
    )(handler)
}

func main() {
    // Setup logging
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

    // Setup database connection
    db, err := setupDB()
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

    // Test database connection
    checkDatabaseConnection(db)

    // Initialize repository and handler
    vehicleRepo := repository.NewVehicleRepository(db)
    vehicleHandler := handlers.NewVehicleHandler(vehicleRepo)

    // Setup routes
    router := mux.NewRouter()
    
    // API routes
    api := router.PathPrefix("/api").Subrouter()
    
    // Vehicle routes
    api.HandleFunc("/vehicles/available", middleware.AuthMiddleware(vehicleHandler.GetAvailableVehicles)).Methods("GET", "OPTIONS")
    
    // Booking routes
    api.HandleFunc("/bookings", middleware.AuthMiddleware(vehicleHandler.CreateBooking)).Methods("POST", "OPTIONS")
    api.HandleFunc("/bookings/{id}", middleware.AuthMiddleware(vehicleHandler.UpdateBooking)).Methods("PUT", "OPTIONS")
    api.HandleFunc("/bookings/{id}", middleware.AuthMiddleware(vehicleHandler.CancelBooking)).Methods("DELETE", "OPTIONS")
    api.HandleFunc("/bookings/my", middleware.AuthMiddleware(vehicleHandler.GetUserBookings)).Methods("GET", "OPTIONS")
    
    // Vehicle status update route
    api.HandleFunc("/vehicles/{id}/status", middleware.AuthMiddleware(vehicleHandler.UpdateVehicleStatus)).Methods("PUT", "OPTIONS")
    

    router.PathPrefix("/").Handler(http.FileServer(http.Dir("frontend")))
    corsHandler := setupCORS(router)

    // Create server
    server := &http.Server{
        Addr:         ":8085",
        Handler:      corsHandler,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start server
    log.Printf("Vehicle Service starting on port 8085\n")
    log.Printf("Serving frontend from: %s\n", filepath.Join(".", "frontend"))
    if err := server.ListenAndServe(); err != nil {
        log.Fatal("Server failed to start:", err)
    }
}