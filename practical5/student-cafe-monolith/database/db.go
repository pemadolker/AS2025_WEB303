package database

import (
	"log"
	"student-cafe-monolith/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(dsn string) error {
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	// Auto-migrate all tables
	err = DB.AutoMigrate(&models.User{}, &models.MenuItem{}, &models.Order{}, &models.OrderItem{})
	if err != nil {
		return err
	}

	log.Println("Database connected and migrated")
	return nil
}
