package models

import (
	"time"
)
type Checkout struct {
	ID 			uint `json:"id" gorm:"primaryKey"`
	CreatedAt 	time.Time
	OrderRefer   int   `json:"order_id"`
	Order 		*Order  `gorm:"foreignKey:OrderRefer"`
	Nominal 	float64  `json:"nominal"`
	Status 		string 	`json:"status"`
}