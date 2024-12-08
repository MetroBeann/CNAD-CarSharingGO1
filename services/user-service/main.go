package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
    "path/filepath"

    gorillaCORS "github.com/gorilla/handlers"
    "github.com/gorilla/mux"
    _ "github.com/lib/pq"

    userHandlers "cnad-carsharinggo/services/user-service/handlers"
    "cnad-carsharinggo/services/user-service/repository"
    "cnad-carsharinggo/services/user-service/middleware"
)

// Configuration constants
const (
    defaultPort = ":8080"
    dbConnRetries = 5
    dbConnRetryDelay = 5 * time.Second
)

func initDB(connStr string) (*sql.DB, error) {
    var db *sql.DB
    var err error

    // Try to connect to the database with retries
    for i := 0; i < dbConnRetries; i++ {
        db, err = sql.Open("postgres", connStr)
        if err != nil {
            log.Printf("Failed to open database connection (attempt %d/%d): %v", i+1, dbConnRetries, err)
            time.Sleep(dbConnRetryDelay)
            continue
        }

        // Test the connection
        err = db.Ping()
        if err == nil {
            log.Println("Successfully connected to the database")
            break
        }

        log.Printf("Failed to ping database (attempt %d/%d): %v", i+1, dbConnRetries, err)
        db.Close()
        time.Sleep(dbConnRetryDelay)
    }

    if err != nil {
        return nil, fmt.Errorf("failed to establish database connection after %d attempts: %v", dbConnRetries, err)
    }

    // Configure connection pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)

    return db, nil
}

func setupRoutes(userHandler *userHandlers.UserHandler) *mux.Router {
    r := mux.NewRouter()

    // API routes
    api := r.PathPrefix("/users").Subrouter()
    api.HandleFunc("/register", userHandler.RegisterUser).Methods("POST", "OPTIONS")
    api.HandleFunc("/login", userHandler.LoginUser).Methods("POST", "OPTIONS")
    api.HandleFunc("/{id}/profile", middleware.AuthMiddleware(userHandler.UpdateUserProfile)).Methods("PUT", "OPTIONS")

    fs := http.FileServer(http.Dir("frontend"))
    r.HandleFunc("/", serveIndex)
    r.PathPrefix("/").Handler(http.StripPrefix("/", fs))

    return r
}

// serveIndex serves the index.html file
func serveIndex(w http.ResponseWriter, r *http.Request) {
    frontendDir := filepath.Join(".", "frontend")
    http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
}

func setupCORS(handler http.Handler) http.Handler {
    return gorillaCORS.CORS(
        gorillaCORS.AllowedOrigins([]string{"*"}),
        gorillaCORS.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
        gorillaCORS.AllowedHeaders([]string{"Content-Type", "Authorization"}),
        gorillaCORS.MaxAge(86400), // 24 hours
    )(handler)
}

func main() {
    // Setup logging
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    
    // Database connection string
    connStr := "postgres://postgres.wjdhhzmaclmsvaiszagk:22KC6282t04@@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres"
    
    // Initialize database connection
    db, err := initDB(connStr)
    if err != nil {
        log.Fatal("Failed to initialize database:", err)
    }
    defer db.Close()

    // Initialize repositories and handlers
    userRepo := repository.NewUserRepository(db)
    userHandler := userHandlers.NewUserHandler(userRepo)

    // Setup routes
    router := setupRoutes(userHandler)

    // Setup CORS
    corsHandler := setupCORS(router)

    // Determine port
    port := os.Getenv("PORT")
    if port == "" {
        port = defaultPort
    }

    // Create server
    server := &http.Server{
        Addr:         port,
        Handler:      corsHandler,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start server
    log.Printf("User Service starting on port %s\n", port)
    log.Printf("Serving frontend from: %s\n", filepath.Join(".", "frontend"))
    if err := server.ListenAndServe(); err != nil {
        log.Fatal("Server failed to start:", err)
    }
}