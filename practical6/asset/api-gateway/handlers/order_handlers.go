package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	orderv1 "github.com/douglasswm/student-cafe-protos/gen/go/order/v1"
	"github.com/go-chi/chi/v5"
)

// CreateOrder handles POST /api/orders
// Translates HTTP request to gRPC CreateOrder call
func (h *Handlers) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Parse HTTP JSON request body
	var req struct {
		UserID uint32 `json:"user_id"`
		Items  []struct {
			MenuItemID uint32 `json:"menu_item_id"`
			Quantity   uint32 `json:"quantity"`
		} `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Convert HTTP items to gRPC OrderItemRequest protobuf messages
	var items []*orderv1.OrderItemRequest
	for _, item := range req.Items {
		items = append(items, &orderv1.OrderItemRequest{
			MenuItemId: item.MenuItemID,
			Quantity:   int32(item.Quantity),
		})
	}

	// Call gRPC service
	resp, err := h.clients.OrderClient.CreateOrder(context.Background(), &orderv1.CreateOrderRequest{
		UserId: req.UserID,
		Items:  items,
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Return HTTP JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp.Order)
}

// GetOrder handles GET /api/orders/{id}
// Translates HTTP request to gRPC GetOrder call
func (h *Handlers) GetOrder(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid order ID", http.StatusBadRequest)
		return
	}

	// Call gRPC service
	resp, err := h.clients.OrderClient.GetOrder(context.Background(), &orderv1.GetOrderRequest{
		Id: uint32(id),
	})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Return HTTP JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Order)
}

// GetOrders handles GET /api/orders
// Translates HTTP request to gRPC GetOrders call
func (h *Handlers) GetOrders(w http.ResponseWriter, r *http.Request) {
	// Call gRPC service
	resp, err := h.clients.OrderClient.GetOrders(context.Background(), &orderv1.GetOrdersRequest{})

	if err != nil {
		handleGRPCError(w, err)
		return
	}

	// Return HTTP JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Orders)
}
