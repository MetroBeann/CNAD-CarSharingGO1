package middleware

import (
    "context"
    "net/http"
    "strings"

    "cnad-carsharinggo/services/user-service/handlers"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get the Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization token", http.StatusUnauthorized)
            return
        }

        // Check if the header starts with "Bearer "
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            http.Error(w, "Invalid token format", http.StatusUnauthorized)
            return
        }

        tokenString := parts[1]

        // Validate the token
        claims, err := handlers.ValidateToken(tokenString)
        if err != nil {
            http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
            return
        }

        // Add claims to request context
        ctx := context.WithValue(r.Context(), "claims", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}