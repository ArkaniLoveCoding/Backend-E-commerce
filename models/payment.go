package models

import "time"

type Payment struct {
	ID            uint `json:"id" gorm:"primaryKey"`
	CheckoutID    uint	`json:"checkout_id"`
	OrderID       uint	`json:"order_id"`
	Order 		  *Order `gorm:"foreignKey:OrderID"`
	Provider      string	`json:"provider"`
	PaymentMethod string	`json:"payment_method"`
	Amount        int32		`json:"amount"`
	Status        string	`json:"status"`
	ProviderRef   string	`json:"provider_ref"`
	PaymentURL    string	`json:"payment_url"`
	PaidAt        *time.Time
	CreatedAt     time.Time
}
