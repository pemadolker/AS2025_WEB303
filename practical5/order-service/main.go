package main

import (
	"log"
	"net/http"
	"os"
	"order-service/database"
	"order-service/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Order endpoints
	r.Post("/orders", handlers.CreateOrder)
	r.Get("/orders", handlers.GetOrders)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	log.Printf("Order service starting on :%s", port)
	http.ListenAndServe(":"+port, r)
}
