package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"order-service/database"
	grpcclient "order-service/grpc"
	"order-service/models"

	menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
	userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
)

var GrpcClients *grpcclient.Clients

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

	ctx := context.Background()

	// Validate user exists via gRPC
	_, err := GrpcClients.UserClient.GetUser(ctx, &userv1.GetUserRequest{
		Id: uint32(req.UserID),
	})
	if err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	// Create order
	order := models.Order{
		UserID: req.UserID,
		Status: "pending",
	}

	// Validate each menu item and snapshot price via gRPC
	for _, item := range req.Items {
		menuItemResp, err := GrpcClients.MenuClient.GetMenuItem(ctx, &menuv1.GetMenuItemRequest{
			Id: uint32(item.MenuItemID),
		})
		if err != nil {
			http.Error(w, "Menu item not found", http.StatusBadRequest)
			return
		}

		orderItem := models.OrderItem{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
			Price:      menuItemResp.MenuItem.Price,
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
