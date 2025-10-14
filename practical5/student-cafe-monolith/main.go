package main

import (
	"log"
	"net/http"
	"os"
	"student-cafe-monolith/database"
	"student-cafe-monolith/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// Connect to database
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=student_cafe port=5432 sslmode=disable"
	}

	if err := database.Connect(dsn); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// User routes
	r.Post("/api/users", handlers.CreateUser)
	r.Get("/api/users/{id}", handlers.GetUser)

	// Menu routes
	r.Get("/api/menu", handlers.GetMenu)
	r.Post("/api/menu", handlers.CreateMenuItem)

	// Order routes
	r.Post("/api/orders", handlers.CreateOrder)
	r.Get("/api/orders", handlers.GetOrders)

	log.Println("Monolith server starting on :8080")
	http.ListenAndServe(":8080", r)
}
