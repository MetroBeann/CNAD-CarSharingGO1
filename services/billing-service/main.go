package main

import (
    "database/sql"
    "log"
    "net/http"
    "path/filepath"
    "time"

    gorillaCORS "github.com/gorilla/handlers"
    "github.com/gorilla/mux"
    _ "github.com/lib/pq"

    "billing-service/handlers"
    "billing-service/middleware"
    "billing-service/repository"
)

const (
    defaultPort = ":8083"
    dbConnRetries = 5
    dbConnRetryDelay = 5 * time.Second
)

func initDB(connStr string) (*sql.DB, error) {
    var db *sql.DB
    var err error

    for i := 0; i < dbConnRetries; i++ {
        db, err = sql.Open("postgres", connStr)
        if err != nil {
            log.Printf("Failed to open database connection (attempt %d/%d): %v", i+1, dbConnRetries, err)
            time.Sleep(dbConnRetryDelay)
            continue
        }

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
        return nil, err
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)

    return db, nil
}

func setupRoutes(billingHandler *handlers.BillingHandler) *mux.Router {
    r := mux.NewRouter()

    // Auth routes
    auth := r.PathPrefix("/api/auth").Subrouter()
    auth.HandleFunc("/login", billingHandler.Login).Methods("POST", "OPTIONS")
    auth.HandleFunc("/register", billingHandler.Register).Methods("POST", "OPTIONS")
    auth.HandleFunc("/validate", billingHandler.ValidateToken).Methods("GET", "OPTIONS")

    // Billing API routes
    api := r.PathPrefix("/api/billing").Subrouter()
    
    // Public endpoints
    api.HandleFunc("/calculate", billingHandler.CalculateEstimate).Methods("POST")
    
    // Protected endpoints
    api.HandleFunc("/invoices", middleware.AuthMiddleware(billingHandler.CreateInvoice)).Methods("POST")
    api.HandleFunc("/users/{id}/invoices", middleware.AuthMiddleware(billingHandler.GetUserInvoices)).Methods("GET")
    api.HandleFunc("/payment-methods", middleware.AuthMiddleware(billingHandler.AddPaymentMethod)).Methods("POST")
    api.HandleFunc("/invoices/{id}/pay", middleware.AuthMiddleware(billingHandler.ProcessPayment)).Methods("POST")

    // Frontend routes
    fs := http.FileServer(http.Dir("frontend"))
    r.HandleFunc("/", serveIndex)
    r.PathPrefix("/").Handler(http.StripPrefix("/", fs))

    return r
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
    frontendDir := filepath.Join(".", "frontend")
    http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
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
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    
    connStr := "postgres://postgres.wjdhhzmaclmsvaiszagk:22KC6282t04@@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres"
    
    db, err := initDB(connStr)
    if err != nil {
        log.Fatal("Failed to initialize database:", err)
    }
    defer db.Close()

    // Initialize repositories and handlers
    billingRepo := repository.NewBillingRepository(db)
    billingHandler := handlers.NewBillingHandler(billingRepo)

    // Setup routes
    router := setupRoutes(billingHandler)
    corsHandler := setupCORS(router)

    server := &http.Server{
        Addr:         defaultPort,
        Handler:      corsHandler,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    log.Printf("Billing Service starting on port %s\n", defaultPort)
    log.Printf("Serving frontend from: %s\n", filepath.Join(".", "frontend"))
    if err := server.ListenAndServe(); err != nil {
        log.Fatal("Server failed to start:", err)
    }
}