package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"user-service/database"
	grpcserver "user-service/grpc"

	userv1 "github.com/douglasswm/student-cafe-protos/gen/go/user/v1"
	"google.golang.org/grpc"
)

func main() {
	// Connect to dedicated user database
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=user_db port=5432 sslmode=disable"
	}

	if err := database.Connect(dsn); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get gRPC port from environment
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9091"
	}

	// Start listening on TCP port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}

	// Create and register gRPC server
	s := grpc.NewServer()
	userv1.RegisterUserServiceServer(s, grpcserver.NewUserServer())

	log.Printf("User service (gRPC only) starting on :%s", grpcPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("gRPC server failed: %v", err)
	}
}
