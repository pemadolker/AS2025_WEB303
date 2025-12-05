package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	UserID     uint        `json:"user_id"`
	Status     string      `json:"status"` // "pending", "completed"
	OrderItems []OrderItem `json:"order_items" gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
	gorm.Model
	OrderID    uint    `json:"order_id"`
	MenuItemID uint    `json:"menu_item_id"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"` // Snapshot price at order time
}
