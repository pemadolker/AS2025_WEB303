package database

import (
	"log"
	"menu-service/models"

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

	// Only migrate menu-related tables
	err = DB.AutoMigrate(&models.MenuItem{})
	if err != nil {
		return err
	}

	log.Println("Menu database connected")
	return nil
}
