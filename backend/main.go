package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"ticket-system/config"
	"ticket-system/middleware"
	"ticket-system/routes"
)

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Connect to database
	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	mux := http.NewServeMux()

	// Register public routes
	mux.HandleFunc("/health", routes.HealthCheck)
	mux.HandleFunc("/auth/register", routes.Register)
	mux.HandleFunc("/auth/login", routes.Login)

	// Register protected ticket routes
	mux.HandleFunc("/tickets", middleware.AuthMiddleware(routes.TicketListAndCreate))
	mux.HandleFunc("/tickets/", middleware.AuthMiddleware(routes.SingleTicket))

	fmt.Printf("Server listening on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, corsMiddleware(mux)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
