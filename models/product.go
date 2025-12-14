package models

import "time"

type Product struct {
	ID 				uint 	`json:"id" gorm:"primaryKey"`
	CreatedAt 		time.Time
	Name			string `json:"name"`
	Price 			float64 `json:"price"`
	Stock 			int32   `json:"stock"`
	Serialnumber 	string `json:"serial_number"`
	Expired 		string  `json:"expired" gorm:"uniqueIndex"`
	Category    	string  `json:"category"`
	Image 			string 	`json:"image" gorm:"uniqueIndex"`
	Status 		 	string  `json:"status"`
}