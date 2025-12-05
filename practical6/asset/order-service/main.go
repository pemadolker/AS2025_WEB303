package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"order-service/database"
	grpcserver "order-service/grpc"

	orderv1 "github.com/douglasswm/student-cafe-protos/gen/go/order/v1"
	"google.golang.org/grpc"
)

func main() {
	// Connect to dedicated order database
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=order_db port=5432 sslmode=disable"
	}

	if err := database.Connect(dsn); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get gRPC port from environment
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9093"
	}

	// Start listening on TCP port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}

	// Get service addresses for gRPC clients (order service calls user and menu)
	userServiceAddr := os.Getenv("USER_SERVICE_GRPC_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = "user-service:9091"
	}

	menuServiceAddr := os.Getenv("MENU_SERVICE_GRPC_ADDR")
	if menuServiceAddr == "" {
		menuServiceAddr = "menu-service:9092"
	}

	// Create order gRPC server with clients to other services
	orderServer, err := grpcserver.NewOrderServer(userServiceAddr, menuServiceAddr)
	if err != nil {
		log.Fatalf("Failed to create gRPC order server: %v", err)
	}

	// Create and register gRPC server
	s := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(s, orderServer)

	log.Printf("Order service (gRPC only) starting on :%s", grpcPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}
