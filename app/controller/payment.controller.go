package controller

import (
	"errors"
	"fmt"
	"strconv"
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
	CheckoutID    int	`json:"checkout_id"`
	OrderID       int	`json:"order_id"`
	Order 		  *Order `json:"order"`
	Provider      string	`json:"provider"`
	PaymentMethod string	`json:"payment_method"`
	Amount        int32		`json:"amount"`	
	Status        string	`json:"status"`
	ProviderRef   time.Time
	PaymentURL    string	`json:"payment_url"`
	PaidAt        time.Time
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
		PaidAt:	payment.PaidAt,
		CreatedAt: payment.CreatedAt,
	}
}

func CreateNewPayment (c *fiber.Ctx) error {
	type ParamsCreate struct {
		ID            uint	`json:"id" gorm:"primaryKey"`
		CheckoutID    int	`json:"checkout_id" validate:"required"`
		OrderID       int	`json:"order_id" validate:"required"`
		Order 		  *Order `json:"order"`
		Provider      string	`json:"provider"`
		PaymentMethod string	`json:"payment_method" validate:"required"`
		Amount        int32		`json:"amount"`	
		Status        string	`json:"status"`
		ProviderRef   string	`json:"provider_ref"`
		PaymentURL    string	`json:"payment_url"`
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

	var orders models.Order
	if err := database.Database.DB.
	Select("id", "total_order", "quantity", "user_refer", "product_refer").
	Where("id = ?", int(body.OrderID)).
	Preload("Product").
	Preload("User").
	First(&orders).Error; err != nil {
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Order tidak ditemukan!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	var checkouts models.Checkout
	if err := database.Database.DB.
	Select("id", "order_id", "nominal", "status").
	Where("id = ?", int(body.CheckoutID)).
	First(&checkouts).Error; err != nil {
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Checkout id tidak ada!")
		}
		return  utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	MidtransResp, err := utils.CreateSnapMidtrans(int(checkouts.Nominal), int(body.OrderID))
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	result := models.Payment{
		CheckoutID: int(body.CheckoutID),
		OrderID: int(body.OrderID),
		Order: &orders,
		Provider: "midtrans",
		PaymentMethod: body.PaymentMethod,
		Amount: int32(checkouts.Nominal),
		Status: "pending",
		ProviderRef: time.Now(),
		PaymentURL: MidtransResp.RedirectURL,
	}

	tx := database.Database.DB.Begin()
	defer func(){
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Create(&result).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal membuat data payment baru!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	if err := tx.
	Select("id", "stock").
	Model(&models.Product{}).
	Where("id = ?", orders.ProductRefer).
	Update("stock", gorm.Expr("stock - ?", orders.Quantity)).Error; err != nil {
		tx.Rollback()
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseUser := CreateUserResponse(*orders.User)
	responeProduct := CreateProductResponse(*orders.Product)
	responseOrder := ResponseToOrder(orders, responseUser, responeProduct)
	responsePayment := DatabaseIntoPayment(result, responseOrder)

	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responsePayment, fiber.StatusOK, "Berhasil membuat payment baru!")
}
func GetOrderId (id int, payments *models.Payment) error {
	if err := database.Database.DB.
	Select("order_id", "provider", "payment_method", "amount", "status", "payment_url").
	Where("order_id = ?").
	First(&payments).Error; err != nil {
		return nil
	}
	if id == 0 {
		return errors.New("Tidak ada id yang ditemukan!")
	}
	return nil
}
func WebHookForPayments (c *fiber.Ctx) error {
	payload := map[string]interface{}{}

	if err := c.BodyParser(&payload); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memvalidasi body!")
	}

	order_id, ok:= payload["order_id"].(string)
	if !ok {
		order_id = ""
	}

	checkout_id, ok := payload["checkout_id"].(string)
	if !ok {
		checkout_id = ""
	}
	convertCheckoutId, err := strconv.Atoi(checkout_id)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	convert, err := strconv.Atoi(order_id)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
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
	if err := FindIdOrder(convert, &orders); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	var payments models.Payment
	if err := GetOrderId(convert, &payments); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	if err := database.Database.DB.
	Select("checkout_id", "order_id", "provider", "payment_method", "amount", "status", "payment_url").
	Where("checkout_id = ?", convertCheckoutId).
	First(&payments).Error; err != nil {
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Checkout id tidak ditemukan!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	paid, err := time.Parse(time.RFC3339, transaction_time)
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
			database.Database.DB.
			Select("id", "status").
			Model(&orders).
			Where("id = ?", order_id).
			Updates(map[string]interface{}{
				"status": "success!",
			}).Error; err != nil {
				if errors.Is(err, gorm.ErrInvalidField) {
					return utils.JsonWithError(c, fiber.StatusBadRequest, "Order id tidak ditemukan!")
				}
				return utils.JsonWithError(c, fiber.StatusBadGateway, err.Error())
			}
			if err := database.Database.DB.
			Select("id", "status").
			Model(&models.Checkout{}).
			Where("id = ?", checkout_id).
			Updates(map[string]interface{}{
				"status": "success!",
			}).Error; err != nil {
				if errors.Is(err, gorm.ErrInvalidField) {
					return utils.JsonWithError(c, fiber.StatusBadRequest, "Checkout id tidak ditemukan!")
				}
				return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
			}
		case "expire":
			dbPaymentStatus = "failed! because the transactions has been expired.."
			if err := 
			database.Database.DB.
			Select("id", "status").
			Model(&orders).
			Where("id = ?", order_id).
			Updates(map[string]interface{}{
				"status": "expired!",
			}).Error; err != nil {
				if errors.Is(err, gorm.ErrInvalidField) {
					return utils.JsonWithError(c, fiber.StatusBadRequest, "Order id tidak ditemukan!")
				}
				return utils.JsonWithError(c, fiber.StatusBadGateway, err.Error())
			}
			if err := database.Database.DB.
			Select("id", "status").
			Model(&models.Checkout{}).
			Where("id = ?").
			Updates(map[string]interface{}{
				"status": "expired!",
			}).Error; err != nil {
				if errors.Is(err, gorm.ErrInvalidField) {
					return utils.JsonWithError(c, fiber.StatusBadRequest, "Checkout id tidak ditemukan!")
				}
				return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
			}
		case "capture":
			dbPaymentStatus = "success to paid!"
			if err := 
			database.Database.DB.
			Select("id", "status").
			Model(&orders).
			Where("id = ?", order_id).
			Updates(map[string]interface{}{
				"status": "success!",
			}).Error; err != nil {
				if errors.Is(err, gorm.ErrInvalidField) {
					return utils.JsonWithError(c, fiber.StatusBadRequest, "Order id tidak ditemukan!")
				}
				return utils.JsonWithError(c, fiber.StatusBadGateway, err.Error())
			}
			if err := database.Database.DB.Model(&models.Checkout{}).
			Where("id = ?").
			Updates(map[string]interface{}{
				"status": "capture",
			}).Error; err != nil {
				if errors.Is(err, gorm.ErrInvalidField) {
					return utils.JsonWithError(c, fiber.StatusBadRequest, "Checkout id tidak ditemukan!")
				}
				return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
			}
		case "cancel":
			dbPaymentStatus = "failed!"
			if err := 
			database.Database.DB.
			Model(&orders).
			Select("id", "status").
			Where("id = ?", order_id).
			Updates(map[string]interface{}{
				"status": "failed!",
			}).Error; err != nil {
				if errors.Is(err, gorm.ErrInvalidField) {
					return utils.JsonWithError(c, fiber.StatusBadRequest, "Order id tidak ditemukan!")
				}
				return utils.JsonWithError(c, fiber.StatusBadGateway, err.Error())
			}
			if err := database.Database.DB.
			Select("id", "status").
			Model(&models.Checkout{}).
			Where("id = ?").
			Updates(map[string]interface{}{
				"status": "canceled!",
			}).Error; err != nil {
				if errors.Is(err, gorm.ErrInvalidField) {
					return utils.JsonWithError(c, fiber.StatusBadRequest, "Checkout id tidak ditemukan!")
				}
				return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
			} 
		case "deny":
			dbPaymentStatus = "failed!"
			if err := 
			database.Database.DB.
			Select("id", "status").
			Model(&orders).
			Where("id = ?", order_id).
			Updates(map[string]interface{}{
				"status": "failed!",
			}).Error; err != nil {
				if errors.Is(err, gorm.ErrInvalidField) {
					return utils.JsonWithError(c, fiber.StatusBadRequest, "Order id tidak ditemukan!")
				}
				return utils.JsonWithError(c, fiber.StatusBadGateway, err.Error())
			}
			if err := database.Database.DB.
			Select("id", "status").
			Model(&models.Checkout{}).
			Where("id = ?").
			Updates(map[string]interface{}{
				"status": "failed!",
			}).Error; err != nil {
				if errors.Is(err, gorm.ErrInvalidField) {
					return utils.JsonWithError(c, fiber.StatusBadRequest, "Checkout id tidak ditemukan!")
				}
				return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
			}
	}

	if err := database.Database.DB.
	Select("order_id", "status", "paid_at", "payment_method").
	Model(&payments).
	Where("order_id = ?", convert).
	Updates(map[string]interface{}{
		"status": dbPaymentStatus,
		"paid_at": paid,
		"payment_method": payment_type,
	}).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return c.Status(200).JSON(fiber.Map{
		"status": "success",
		"isValid": "true",
		"message": "Berhasil!",
	})
}