package database

import (
	"log"
	"order-service/models"

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

	// Only migrate order-related tables
	err = DB.AutoMigrate(&models.Order{}, &models.OrderItem{})
	if err != nil {
		return err
	}

	log.Println("Order database connected")
	return nil
}
