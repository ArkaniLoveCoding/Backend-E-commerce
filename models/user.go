package models

import "time"

type User struct {
	ID 			uint `json:"id" gorm:"primaryKey"`
	CreatedAt 	time.Time
	Name 		string `json:"name" gorm:"uniqueIndex"`
	Password 	string `json:"password" gorm:"uniqueIndex"`
	Email 		string `json:"email" gorm:"uniqueIndex"`
	Role 		string `json:"role"`
}	