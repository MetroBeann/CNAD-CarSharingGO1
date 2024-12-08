// Path: services/user-service/main.go
package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"

    "github.com/gorilla/mux"
    _ "github.com/lib/pq"

)

func main() {
    // Database connection
    connStr := "postgres://postgres.wjdhhzmaclmsvaiszagk:22KC6282t04@@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal("Database connection error:", err)
    }
    defer db.Close()


    // Setup router
    r := mux.NewRouter()
    
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./frontend/"))))

    // Start server
    port := ":8080"
    fmt.Printf("User Service running on port %s\n", port)
    log.Fatal(http.ListenAndServe(port, r))
}