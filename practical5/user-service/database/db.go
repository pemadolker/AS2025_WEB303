package database

import (
	"log"
	"user-service/models"

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

	// Only migrate user-related tables
	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		return err
	}

	log.Println("User database connected")
	return nil
}
