
package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "practical-one/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct {
	pb.UnimplementedGreeterServiceServer
	timeClient pb.TimeServiceClient
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloResponse, error) {
	timeRes, err := s.timeClient.GetTime(ctx, &pb.TimeRequest{})
	if err != nil {
		return nil, err
	}
	message := fmt.Sprintf("Hello %s! The current time is %s", in.Name, timeRes.CurrentTime)
	return &pb.HelloResponse{Message: message}, nil
}

func main() {
	conn, err := grpc.Dial("time-service:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to time-service: %v", err)
	}
	defer conn.Close()

	timeClient := pb.NewTimeServiceClient(conn)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServiceServer(s, &server{timeClient: timeClient})
	log.Printf("Greeter service listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
