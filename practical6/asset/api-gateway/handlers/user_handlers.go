package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
	"github.com/go-chi/chi/v5"
)

// CreateUser handles POST /api/users
// Translates HTTP request to gRPC CreateUser call
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Parse HTTP JSON request body
	var req struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		IsCafeOwner bool   `json:"is_cafe_owner"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Call gRPC service
	resp, err := h.clients.UserClient.CreateUser(context.Background(), &userv1.CreateUserRequest{
		Name:        req.Name,
		Email:       req.Email,
		IsCafeOwner: req.IsCafeOwner,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Return HTTP JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp.User)
}

// GetUser handles GET /api/users/{id}
// Translates HTTP request to gRPC GetUser call
func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}

	// Call gRPC service
	resp, err := h.clients.UserClient.GetUser(context.Background(), &userv1.GetUserRequest{
		Id: uint32(id),
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Return HTTP JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.User)
}

// GetUsers handles GET /api/users
// Translates HTTP request to gRPC GetUsers call
func (h *Handlers) GetUsers(w http.ResponseWriter, r *http.Request) {
	// Call gRPC service
	resp, err := h.clients.UserClient.GetUsers(context.Background(), &userv1.GetUsersRequest{})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Return HTTP JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Users)
}
