package controller

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ArkaniLoveCoding/fiber-project/database"
	"github.com/ArkaniLoveCoding/fiber-project/models"
	"github.com/ArkaniLoveCoding/fiber-project/utils"
)


type Payment struct {
	ID            uint	`json:"id" gorm:"primaryKey"`
	CheckoutID    uint	`json:"checkout_id"`
	OrderID       uint	`json:"order_id"`
	Order 		  *Order `json:"order"`
	Provider      string	`json:"provider"`
	PaymentMethod string	`json:"payment_method"`
	Amount        int32		`json:"amount"`	
	Status        string	`json:"status"`
	ProviderRef   string	`json:"provider_ref"`
	PaymentURL    string	`json:"payment_url"`
	PaidAt        *time.Time
	CreatedAt     time.Time
}

func DatabaseIntoPayment (payment models.Payment, order Order) Payment {
	return Payment{
		ID: payment.ID,
		CheckoutID: payment.CheckoutID,
		OrderID: payment.OrderID,
		Order: &order,
		Provider: payment.Provider,
		PaymentMethod: payment.PaymentMethod,
		Amount: payment.Amount,
		Status: payment.Status,
		ProviderRef: payment.ProviderRef,
		PaymentURL: payment.PaymentURL,
		PaidAt: &time.Time{},
		CreatedAt: payment.CreatedAt,
	}
}

func CreateNewPayment (c *fiber.Ctx) error {
	type ParamsCreate struct {
		ID            uint	`json:"id" gorm:"primaryKey"`
		CheckoutID    uint	`json:"checkout_id" validate:"required"`
		OrderID       uint	`json:"order_id" validate:"required"`
		Order 		  *Order `json:"order"`
		Provider      string	`json:"provider"`
		PaymentMethod string	`json:"payment_method" validate:"required"`
		Amount        int32		`json:"amount"`	
		Status        string	`json:"status"`
		ProviderRef   string	`json:"provider_ref"`
		PaymentURL    string	`json:"payment_url"`
		PaidAt        *time.Time
		CreatedAt     time.Time
	}
	var body ParamsCreate
	if err := c.BodyParser(&body); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memverifikasi body!")
	}

		// pengecekan validator
	var validate *validator.Validate = validator.New()
	if err := validate.Struct(body); err != nil {
		var errs []string
		for _, e := range err.(validator.ValidationErrors) {
			errs = append(errs, fmt.Sprintf("%s is %s", e.Field(), e.Tag()))
		}
		return c.Status(400).JSON(fiber.Map{
			"error":   true,
			"message": "Validasi gagal",
			"details": errs,
		})
	}


	var checkouts models.Checkout
	if err := FindIdCheckout(int(body.CheckoutID), &checkouts); err != nil {
		return  utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	var orders models.Order
	if err := FindIdOrder(int(body.OrderID), &orders); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	var products models.Product
	if err := findID(orders.ProductRefer, &products); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id product!")
	}

	var users models.User
	if err := FindIdUser(orders.UserRefer, &users); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id user!")
	}

	times := fmt.Sprintf("Dibayar pada tanggal - %d", time.Now().Unix())
	MidtransResp, err := utils.CreateSnapMidtrans(int(checkouts.Nominal), int(body.OrderID))
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}
	result := models.Payment{
		ID: body.ID,
		CheckoutID: body.CheckoutID,
		OrderID: body.OrderID,
		Provider: "midtrans",
		PaymentMethod: body.PaymentMethod,
		Amount: int32(checkouts.Nominal),
		Status: "pending",
		ProviderRef: times,
		PaymentURL: MidtransResp.RedirectURL,
	}

	tx := database.Database.DB.Begin()
	defer func(){
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&result).Error; err != nil {
		tx.Rollback()
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	if err := tx.Model(&products).Where("id = ?", orders.ProductRefer).Update("stock", gorm.Expr("stock - ?", - orders.Quantity)).Error; err != nil {
		tx.Rollback()
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}
	responsePayment := DatabaseIntoPayment(result, ResponseToOrder(orders, *body.Order.User, *body.Order.Product))
	return utils.JsonWithSuccess(c, responsePayment, fiber.StatusOK, "Berhasil membuat payment!")
}
func WebHookForPayments (c *fiber.Ctx) error {
	payload := map[string]interface{}{}

	if err := c.BodyParser(&payload); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memvalidasi body!")
	}

	order_id, ok:= payload["order_id"].(int)
	if !ok {
		order_id = 0
	}

	transaction_status, ok := payload["transaction_status"].(string)
	if !ok {
		transaction_status = ""
	}

	payment_type, ok := payload["payment_type"].(string)
	if !ok {
		payment_type = ""
	}

	transaction_time, ok := payload["transaction_time"].(string)
	if !ok {
		transaction_time = ""
	}

	var orders models.Order
	if err := database.Database.DB.Find(&orders, "id = ?", order_id).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	var payments models.Payment
	if err := database.Database.DB.Find(&payments, "order_id = ?", order_id).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	paid, err := time.Parse("2006-01-02 15:04:05", transaction_time)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}


	dbPaymentStatus := "pending"

	switch transaction_status {
		case "pending":
			dbPaymentStatus = "pending"
		case "settlement":
			dbPaymentStatus = "success to paid!"
			if err := 
			database.Database.DB.Model(&orders).
			Where("id = ?", order_id).
			Updates(map[string]interface{}{
				"status": "success!",
			}).Error; err != nil {
				return utils.JsonWithError(c, fiber.StatusBadGateway, err.Error())
			}
		case "expire":
			dbPaymentStatus = "failed! because the transactions has been expired.."
			if err := 
			database.Database.DB.Model(&orders).
			Where("id = ?", order_id).
			Updates(map[string]interface{}{
				"status": "expired!",
			}).Error; err != nil {
				return utils.JsonWithError(c, fiber.StatusBadGateway, err.Error())
			}
		case "capture":
			dbPaymentStatus = "success to paid!"
			if err := 
			database.Database.DB.Model(&orders).
			Where("id = ?", order_id).
			Updates(map[string]interface{}{
				"status": "success!",
			}).Error; err != nil {
				return utils.JsonWithError(c, fiber.StatusBadGateway, err.Error())
			}
		case "cancel":
			dbPaymentStatus = "failed!"
			if err := 
			database.Database.DB.Model(&orders).
			Where("id = ?", order_id).
			Updates(map[string]interface{}{
				"status": "failed!",
			}).Error; err != nil {
				return utils.JsonWithError(c, fiber.StatusBadGateway, err.Error())
			}
		case "deny":
			dbPaymentStatus = "failed!"
			if err := 
			database.Database.DB.Model(&orders).
			Where("id = ?", order_id).
			Updates(map[string]interface{}{
				"status": "failed!",
			}).Error; err != nil {
				return utils.JsonWithError(c, fiber.StatusBadGateway, err.Error())
			}
	}

	if err := database.Database.DB.Model(&payments).Updates(map[string]interface{}{
		"status": dbPaymentStatus,
		"paid": paid,
		"payment_method": payment_type,
	}).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return c.SendStatus(200)
}