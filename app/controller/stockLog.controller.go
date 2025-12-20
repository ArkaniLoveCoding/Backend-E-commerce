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

type StockLog struct {
	ID 			uint `json:"id" gorm:"primaryKey"`
	ProductRefer int `json:"product_id"`
	Product 	*Product `gorm:"foreignKey:ProductRefer"`
	PaymentRefer int 	`json:"payment_id"`
	Payment 	*Payment	`json:"payment"`
	OrderRefer 	int 	`json:"order_id"`
	Order 		*Order	`json:"order"`
	CreatedAt 	time.Time
	Note 		string 	`json:"note"`
	Change 		int 	`json:"change"`
	OldStock	int 	`json:"old_stock"`
	NewStock 	int	`json:"new_stock"`
	Type 		string 	`json:"type"`	
}

func DataBaseIntoStock (stock models.StockLog, product Product, payment Payment, order Order) StockLog {
	return StockLog{
		ID: stock.ID,
		ProductRefer: stock.ProductRefer,
		Product: &product,
		PaymentRefer: stock.PaymentRefer,
		Payment: &payment,
		OrderRefer: stock.OrderRefer,
		Order: &order,
		CreatedAt: time.Now(),
		Note: stock.Note,
		Change: stock.Change,
		OldStock: stock.OldStock,
		NewStock: stock.NewStock,
		Type: stock.Type,
	}
}
func CreateNewNote (c *fiber.Ctx) error {
	type ParamsCreate struct {
		ID 			uint `json:"id" gorm:"primaryKey"`
		ProductRefer int 	`json:"product_id"`
		Product 	*Product `json:"product"`
		PaymentRefer int 	`json:"payment_id"`
		Payment 	*Payment	`json:"payment"`
		OrderRefer 	int 	`json:"order_id"`
		Order 		*Order	`json:"order"`
		CreatedAt 	time.Time
		Note 		string 	`json:"note"`
		Change 		int 	`json:"change"`
		OldStock	int 	`json:"old_stock"`
		NewStock 	int 	`json:"new_stock"`
		Type 		string 	`json:"type"`	
	}
	var params ParamsCreate
	if err := c.BodyParser(&params); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengkonversi data!")
	}

		// pengecekan validator
	var validate *validator.Validate = validator.New()
	if err := validate.Struct(params); err != nil {
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

	var payment models.Payment
	if err := database.Database.DB.
	Find(&payment, "id = ?", int(params.PaymentRefer)).Error; err != nil {
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Tidak menemukan Payment id!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	var orders models.Order
	if err := database.Database.DB.
	Preload("Product").
	Preload("User").
	First(&orders, int(params.OrderRefer)).Error; err != nil {
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Tidak menemukan order id!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	if orders.Product == nil {
		fmt.Println(orders.Product)
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Tidak menemukan id product!")
	}
	if orders.User == nil {
		fmt.Println(orders.User)
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Tidak menemukan id user!")
	}
	if payment.ID == 0 {
		fmt.Println(payment.ID)
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Tidak menemukan id payment")
	}

	if err := database.Database.DB.
	Where("status = ?", "success to paid!").
	First(&payment, int(params.PaymentRefer)).Error; err != nil {
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Status tidak ditemukan!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	// main logic

	var products models.Product
	if err := database.Database.DB.
	Find(products, "id = ?", params.ProductRefer).Error; err != nil {
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Tidak menemukan product id!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	var stock models.StockLog
	oldStock := products.Stock
	newStock := products.Stock - orders.Quantity

	products.Stock = newStock
	note := fmt.Sprintf("Order dengan id %d, Stock yang awalnya %d, berubah menjadi %d", orders.ID, oldStock, newStock)
	change := fmt.Sprintf("-%d", orders.Quantity)
	changeFinal, err := strconv.Atoi(change)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	result := models.StockLog{
		ID: params.ID,
		PaymentRefer: params.PaymentRefer,
		Payment: &payment,
		OrderRefer: params.OrderRefer,
		Order: &orders,
		Note: note,
		Change: changeFinal,
		OldStock: int(oldStock),
		NewStock: int(newStock),
		Type: "Order",
	}

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Create(&result).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidValue) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Value tidak match!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	// konfigurasi beberapa response sesuai kebutuhan;
	
	responseUserOrders := CreateUserResponse(*orders.User)
	responseProductForOrders := CreateProductResponse(*orders.Product)
	responeProductForStock := CreateProductResponse(products)
	responseOrder := ResponseToOrder(orders, responseUserOrders, responseProductForOrders)
	responsePayment := DatabaseIntoPayment(payment, responseOrder)
	responseStockLog := DataBaseIntoStock(stock, responeProductForStock, responsePayment, responseOrder)

	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseStockLog, fiber.StatusOK, "Berhasil membuat catatan order!")
}



