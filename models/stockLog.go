package models

import (
	"time"
)

type StockLog struct {
	ID 			uint `json:"id" gorm:"primaryKey"`
	ProductRefer int `json:"product_id"`
	Product 	*Product `gorm:"foreignKey:ProductRefer"`
	PaymentRefer int 	`json:"payment_id"`
	Payment 	*Payment	`gorm:"foreignKey:PaymentRefer"`
	OrderRefer 	int 	`json:"order_id"`
	Order 		*Order	`gorm:"foreignKey:OrderRefer"`
	CreatedAt 	time.Time
	Note 		string 	`json:"note"`
	Change 		int 	`json:"change"`
	OldStock 	int   `json:"old_stock"`
	NewStock 	int 	`json:"new_stock"`
	Type 		string 	`json:"type"`	
}