package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ticket-system/config"
	"ticket-system/middleware"
	"ticket-system/models"
)

// TicketListAndCreate handles GET /tickets and POST /tickets
func TicketListAndCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	// Create ticket
	if r.Method == http.MethodPost {
		var req models.CreateTicketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		req.Title = strings.TrimSpace(req.Title)
		req.Description = strings.TrimSpace(req.Description)

		if req.Title == "" || req.Description == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Title and description required"})
			return
		}

		now := time.Now()
		var ticketID int64
		err = config.DB.QueryRow("INSERT INTO tickets (title, description, status, user_id, created_at, updated_at) VALUES ($1, $2, 'open', $3, $4, $5) RETURNING id", req.Title, req.Description, userID, now, now).Scan(&ticketID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create ticket"})
			return
		}

		ticket := models.Ticket{
			ID:          ticketID,
			Title:       req.Title,
			Description: req.Description,
			Status:      "open",
			UserID:      userID,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ticket)
		return
	}

	// List tickets for logged-in user
	if r.Method == http.MethodGet {
		rows, err := config.DB.Query("SELECT id, title, description, status, user_id, created_at, updated_at FROM tickets WHERE user_id = $1 ORDER BY id DESC", userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch tickets"})
			return
		}
		defer rows.Close()

		tickets := []models.Ticket{}
		for rows.Next() {
			var t models.Ticket
			rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.UserID, &t.CreatedAt, &t.UpdatedAt)
			tickets = append(tickets, t)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tickets)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
}

// SingleTicket handles GET /tickets/{id} and PATCH /tickets/{id}/status
func SingleTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserIDFromContext(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/tickets/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid ticket ID"})
		return
	}

	ticketID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid ticket ID"})
		return
	}

	var t models.Ticket
	err = config.DB.QueryRow("SELECT id, title, description, status, user_id, created_at, updated_at FROM tickets WHERE id = $1", ticketID).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.UserID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Ticket not found"})
		return
	}

	// Ownership check: user can only access own ticket
	if t.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access denied. You do not own this ticket"})
		return
	}

	// GET /tickets/{id}
	if r.Method == http.MethodGet && len(parts) == 1 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(t)
		return
	}

	// PATCH /tickets/{id}/status
	if r.Method == http.MethodPatch && len(parts) == 2 && parts[1] == "status" {
		var req models.UpdateStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		newStatus := strings.TrimSpace(strings.ToLower(req.Status))

		// Validate status sequence: open -> in_progress -> closed
		if t.Status == "closed" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "A closed ticket cannot be reopened"})
			return
		}

		if newStatus != "open" && newStatus != "in_progress" && newStatus != "closed" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid status"})
			return
		}

		if t.Status == "open" && (newStatus != "in_progress" && newStatus != "closed") {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid status transition"})
			return
		}

		if t.Status == "in_progress" && newStatus != "closed" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid status transition"})
			return
		}

		now := time.Now()
		config.DB.Exec("UPDATE tickets SET status = $1, updated_at = $2 WHERE id = $3", newStatus, now, ticketID)

		t.Status = newStatus
		t.UpdatedAt = now
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(t)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
}
