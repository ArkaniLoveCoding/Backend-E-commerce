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
	CreatedAt 	time.Time
	Note 		string 	`json:"note"`
	Change 		int32 	`json:"change"`
	OldStock 	int32   `json:"old_stock"`
	NewStock 	int32 	`json:"new_stock"`
	Type 		string 	`json:"type"`	
}