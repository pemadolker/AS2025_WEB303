package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "practical-one/proto/gen"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTimeServiceServer
}

func (s *server) GetTime(ctx context.Context, in *pb.TimeRequest) (*pb.TimeResponse, error) {
	currentTime := time.Now().Format(time.RFC3339)
	return &pb.TimeResponse{CurrentTime: currentTime}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterTimeServiceServer(s, &server{})
	log.Printf("Time service listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
