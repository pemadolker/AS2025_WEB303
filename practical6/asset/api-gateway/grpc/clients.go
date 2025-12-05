package grpc

import (
	"fmt"
	"log"
	"os"

	menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
	orderv1 "github.com/douglasswm/student-cafe-protos/gen/go/order/v1"
	userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceClients holds all gRPC clients for backend services
type ServiceClients struct {
	UserClient  userv1.UserServiceClient
	MenuClient  menuv1.MenuServiceClient
	OrderClient orderv1.OrderServiceClient
}

// NewServiceClients creates and initializes gRPC clients for all backend services
func NewServiceClients() (*ServiceClients, error) {
	// Get service addresses from environment or use defaults
	userAddr := getEnv("USER_SERVICE_GRPC_ADDR", "user-service:9091")
	menuAddr := getEnv("MENU_SERVICE_GRPC_ADDR", "menu-service:9092")
	orderAddr := getEnv("ORDER_SERVICE_GRPC_ADDR", "order-service:9093")

	log.Printf("Connecting to User Service at %s", userAddr)
	// Create gRPC connection to user service
	userConn, err := grpc.NewClient(userAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	log.Printf("Connecting to Menu Service at %s", menuAddr)
	// Create gRPC connection to menu service
	menuConn, err := grpc.NewClient(menuAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to menu service: %w", err)
	}

	log.Printf("Connecting to Order Service at %s", orderAddr)
	// Create gRPC connection to order service
	orderConn, err := grpc.NewClient(orderAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}

	return &ServiceClients{
		UserClient:  userv1.NewUserServiceClient(userConn),
		MenuClient:  menuv1.NewMenuServiceClient(menuConn),
		OrderClient: orderv1.NewOrderServiceClient(orderConn),
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
