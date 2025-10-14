package handlers

import (
	"encoding/json"
	"net/http"
	"student-cafe-monolith/database"
	"student-cafe-monolith/models"
)

type CreateOrderRequest struct {
	UserID uint `json:"user_id"`
	Items  []struct {
		MenuItemID uint `json:"menu_item_id"`
		Quantity   int  `json:"quantity"`
	} `json:"items"`
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate user exists
	var user models.User
	if err := database.DB.First(&user, req.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	// Create order
	order := models.Order{
		UserID: req.UserID,
		Status: "pending",
	}

	// Build order items
	for _, item := range req.Items {
		var menuItem models.MenuItem
		if err := database.DB.First(&menuItem, item.MenuItemID).Error; err != nil {
			http.Error(w, "Menu item not found", http.StatusBadRequest)
			return
		}

		orderItem := models.OrderItem{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
			Price:      menuItem.Price, // Snapshot current price
		}
		order.OrderItems = append(order.OrderItems, orderItem)
	}

	if err := database.DB.Create(&order).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func GetOrders(w http.ResponseWriter, r *http.Request) {
	var orders []models.Order
	if err := database.DB.Preload("OrderItems").Find(&orders).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
