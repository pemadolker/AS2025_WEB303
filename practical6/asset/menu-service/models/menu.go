package models

import "gorm.io/gorm"

type Menu struct {
	gorm.Model
	Name        string  `json:"name"`
	Description string  `json:"description"`
	MenuItems   []MenuItem `json:"menu_items" gorm:"references:MenuID"`
}

type MenuItem struct {
	gorm.Model
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}
