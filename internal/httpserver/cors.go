package httpserver

import (
    "net/http"
)

// CORS is a minimal middleware that enables cross-origin requests
// from Swagger UI (localhost:8089) and other origins during development.
func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Allow all origins in dev; tighten if needed.
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Accept-Language, Content-Language")
        w.Header().Set("Access-Control-Max-Age", "600")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}

