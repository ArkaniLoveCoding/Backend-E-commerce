package models

import "time"

type Order struct {
	ID 			uint `json:"id" gorm:"primaryKey"`
	CreatedAt 	time.Time
	Order 		time.Time
	Quantity  	int32 `json:"quantity"`
	Status 		string `json:"status"`
	TotalOrder 	float64 `json:"total_order"`

	ProductRefer int `json:"product_id"`
	Product 	*Product `gorm:"foreignKey:ProductRefer"`
	UserRefer 	int `json:"user_id"`
	User 		*User `gorm:"foreignKey:UserRefer"`
}