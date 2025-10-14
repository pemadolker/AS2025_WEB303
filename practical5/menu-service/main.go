package main

import (
	"log"
	"net/http"
	"os"
	"menu-service/database"
	"menu-service/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Menu endpoints (note: no /api prefix)
	r.Get("/menu", handlers.GetMenu)
	r.Post("/menu", handlers.CreateMenuItem)
	r.Get("/menu/{id}", handlers.GetMenuItem)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Menu service starting on :%s", port)
	http.ListenAndServe(":"+port, r)
}
