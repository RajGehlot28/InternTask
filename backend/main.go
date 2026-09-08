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
	mux.HandleFunc("/api/health", routes.HealthCheck)

	mux.HandleFunc("/auth/register", routes.Register)
	mux.HandleFunc("/api/auth/register", routes.Register)

	mux.HandleFunc("/auth/login", routes.Login)
	mux.HandleFunc("/api/auth/login", routes.Login)

	// Register protected ticket routes
	mux.HandleFunc("/tickets", middleware.AuthMiddleware(routes.TicketListAndCreate))
	mux.HandleFunc("/api/tickets", middleware.AuthMiddleware(routes.TicketListAndCreate))

	mux.HandleFunc("/tickets/", middleware.AuthMiddleware(routes.SingleTicket))
	mux.HandleFunc("/api/tickets/", middleware.AuthMiddleware(routes.SingleTicket))

	host := os.Getenv("HOST")
	if host == "" {
		if os.Getenv("RENDER") != "" {
			host = "0.0.0.0"
		} else {
			host = "127.0.0.1"
		}
	}

	addr := host + ":" + port
	if host == "0.0.0.0" {
		addr = ":" + port
	}

	fmt.Printf("Server listening on %s...\n", addr)
	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
