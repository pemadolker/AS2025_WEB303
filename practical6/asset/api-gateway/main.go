package main

import (
	"log"
	"net/http"

	"api-gateway/grpc"
	"api-gateway/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Initialize gRPC clients for all backend services
	clients, err := grpc.NewServiceClients()
	if err != nil {
		log.Fatalf("Failed to create gRPC clients: %v", err)
	}
	log.Println("gRPC clients initialized successfully")

	// Create handlers with gRPC clients
	h := handlers.NewHandlers(clients)

	// Setup HTTP router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// User routes - HTTP to gRPC translation
	r.Post("/api/users", h.CreateUser)
	r.Get("/api/users/{id}", h.GetUser)
	r.Get("/api/users", h.GetUsers)

	// Menu routes - HTTP to gRPC translation
	r.Post("/api/menu", h.CreateMenuItem)
	r.Get("/api/menu/{id}", h.GetMenuItem)
	r.Get("/api/menu", h.GetMenu)

	// Order routes - HTTP to gRPC translation
	r.Post("/api/orders", h.CreateOrder)
	r.Get("/api/orders/{id}", h.GetOrder)
	r.Get("/api/orders", h.GetOrders)

	log.Println("API Gateway starting on :8081 (HTTP→gRPC translation layer)")
	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
