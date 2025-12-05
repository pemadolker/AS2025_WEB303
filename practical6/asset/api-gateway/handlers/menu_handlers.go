package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
	"github.com/go-chi/chi/v5"
)

// CreateMenuItem handles POST /api/menu
// Translates HTTP request to gRPC CreateMenuItem call
func (h *Handlers) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	// Parse HTTP JSON request body
	var req struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Call gRPC service
	resp, err := h.clients.MenuClient.CreateMenuItem(context.Background(), &menuv1.CreateMenuItemRequest{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Return HTTP JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp.MenuItem)
}

// GetMenuItem handles GET /api/menu/{id}
// Translates HTTP request to gRPC GetMenuItem call
func (h *Handlers) GetMenuItem(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid menu item ID", http.StatusBadRequest)
		return
	}

	// Call gRPC service
	resp, err := h.clients.MenuClient.GetMenuItem(context.Background(), &menuv1.GetMenuItemRequest{
		Id: uint32(id),
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Return HTTP JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.MenuItem)
}

// GetMenu handles GET /api/menu
// Translates HTTP request to gRPC GetMenu call
func (h *Handlers) GetMenu(w http.ResponseWriter, r *http.Request) {
	// Call gRPC service
	resp, err := h.clients.MenuClient.GetMenu(context.Background(), &menuv1.GetMenuRequest{})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Return HTTP JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.MenuItems)
}
