package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"menu-service/database"
	grpcserver "menu-service/grpc"

	menuv1 "github.com/douglasswm/student-cafe-protos/gen/go/menu/v1"
	"google.golang.org/grpc"
)

func main() {
	// Connect to dedicated menu database
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=menu_db port=5432 sslmode=disable"
	}

	if err := database.Connect(dsn); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get gRPC port from environment
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9092"
	}

	// Start listening on TCP port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}

	// Create and register gRPC server
	s := grpc.NewServer()
	menuv1.RegisterMenuServiceServer(s, grpcserver.NewMenuServer())

	log.Printf("Menu service (gRPC only) starting on :%s", grpcPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}
