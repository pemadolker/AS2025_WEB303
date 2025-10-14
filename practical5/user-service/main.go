package main

import (
	"log"
	"net/http"
	"os"
	"user-service/database"
	"user-service/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// User endpoints
	r.Post("/users", handlers.CreateUser)
	r.Get("/users/{id}", handlers.GetUser)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("User service starting on :%s", port)
	http.ListenAndServe(":"+port, r)
}
