package grpc

import (
	"context"
	"fmt"
	"time"

	menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
	orderv1 "github.com/douglasswm/student-cafe-protos/gen/go/order/v1"
	userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"order-service/database"
	"order-service/models"
)

// OrderServer implements the gRPC OrderService
type OrderServer struct {
	orderv1.UnimplementedOrderServiceServer
	UserClient userv1.UserServiceClient
	MenuClient menuv1.MenuServiceClient
}

// NewOrderServer creates a new gRPC order server
func NewOrderServer(userServiceAddr, menuServiceAddr string) (*OrderServer, error) {
	// Connect to user service
	userConn, err := grpc.NewClient(
		userServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	// Connect to menu service
	menuConn, err := grpc.NewClient(
		menuServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to menu service: %w", err)
	}

	return &OrderServer{
		UserClient: userv1.NewUserServiceClient(userConn),
		MenuClient: menuv1.NewMenuServiceClient(menuConn),
	}, nil
}

// CreateOrder creates a new order
func (s *OrderServer) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	// Validate user exists via gRPC
	_, err := s.UserClient.GetUser(ctx, &userv1.GetUserRequest{Id: req.UserId})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "user not found: %v", err)
	}

	// Create order
	order := models.Order{
		UserID: uint(req.UserId),
		Status: "pending",
	}

	// Validate menu items and snapshot prices via gRPC
	for _, item := range req.Items {
		menuItemResp, err := s.MenuClient.GetMenuItem(ctx, &menuv1.GetMenuItemRequest{Id: item.MenuItemId})
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "menu item %d not found: %v", item.MenuItemId, err)
		}

		orderItem := models.OrderItem{
			MenuItemID: uint(item.MenuItemId),
			Quantity:   int(item.Quantity),
			Price:      menuItemResp.MenuItem.Price,
		}
		order.OrderItems = append(order.OrderItems, orderItem)
	}

	// Save order to database
	if err := database.DB.Create(&order).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}

	return &orderv1.CreateOrderResponse{
		Order: modelToProto(&order),
	}, nil
}

// GetOrders retrieves all orders
func (s *OrderServer) GetOrders(ctx context.Context, req *orderv1.GetOrdersRequest) (*orderv1.GetOrdersResponse, error) {
	var orders []models.Order
	if err := database.DB.Preload("OrderItems").Find(&orders).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get orders: %v", err)
	}

	protoOrders := make([]*orderv1.Order, len(orders))
	for i, order := range orders {
		protoOrders[i] = modelToProto(&order)
	}

	return &orderv1.GetOrdersResponse{
		Orders: protoOrders,
	}, nil
}

// GetOrder retrieves an order by ID
func (s *OrderServer) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	var order models.Order
	if err := database.DB.Preload("OrderItems").First(&order, req.Id).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "order not found")
	}

	return &orderv1.GetOrderResponse{
		Order: modelToProto(&order),
	}, nil
}

// modelToProto converts a GORM Order model to proto Order message
func modelToProto(order *models.Order) *orderv1.Order {
	protoItems := make([]*orderv1.OrderItem, len(order.OrderItems))
	for i, item := range order.OrderItems {
		protoItems[i] = &orderv1.OrderItem{
			Id:         uint32(item.ID),
			OrderId:    uint32(item.OrderID),
			MenuItemId: uint32(item.MenuItemID),
			Quantity:   int32(item.Quantity),
			Price:      item.Price,
			CreatedAt:  item.CreatedAt.Format(time.RFC3339),
			UpdatedAt:  item.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &orderv1.Order{
		Id:         uint32(order.ID),
		UserId:     uint32(order.UserID),
		Status:     order.Status,
		OrderItems: protoItems,
		CreatedAt:  order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  order.UpdatedAt.Format(time.RFC3339),
	}
}
