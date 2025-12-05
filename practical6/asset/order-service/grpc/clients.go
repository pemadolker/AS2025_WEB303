package grpc

import (
	"fmt"
	"os"

	menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
	userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients holds gRPC client connections
type Clients struct {
	UserClient userv1.UserServiceClient
	MenuClient menuv1.MenuServiceClient
}

// NewClients creates gRPC clients for user and menu services
func NewClients() (*Clients, error) {
	// Get service addresses from environment or use defaults
	userServiceAddr := os.Getenv("USER_SERVICE_GRPC_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = "user-service:9091"
	}

	menuServiceAddr := os.Getenv("MENU_SERVICE_GRPC_ADDR")
	if menuServiceAddr == "" {
		menuServiceAddr = "menu-service:9092"
	}

	// Connect to user service
	userConn, err := grpc.NewClient(
		userServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service at %s: %w", userServiceAddr, err)
	}

	// Connect to menu service
	menuConn, err := grpc.NewClient(
		menuServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to menu service at %s: %w", menuServiceAddr, err)
	}

	return &Clients{
		UserClient: userv1.NewUserServiceClient(userConn),
		MenuClient: menuv1.NewMenuServiceClient(menuConn),
	}, nil
}
