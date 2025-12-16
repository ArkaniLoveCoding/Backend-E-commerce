package controller

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/ArkaniLoveCoding/fiber-project/database"
	"github.com/ArkaniLoveCoding/fiber-project/models"
	"github.com/ArkaniLoveCoding/fiber-project/utils"
)

type StockLog struct {
	ID 			uint `json:"id" gorm:"primaryKey"`
	ProductRefer int `json:"product_id"`
	Product 	*Product `json:"product"`
	PaymentRefer int 	`json:"payment_id"`
	Payment 	*Payment	`gorm:"foreignKey:PaymentRefer"`
	CreatedAt 	time.Time
	Note 		string 	`json:"note"`
	Change 		int32 	`json:"change"`
	OldStock	int32 	`json:"old_stock"`
	NewStock 	int32 	`json:"new_stock"`
	Type 		string 	`json:"type"`	
}

func DataBaseIntoStock (stock models.StockLog, product Product, payment Payment) StockLog {
	return StockLog{
		ID: stock.ID,
		ProductRefer: stock.ProductRefer,
		Product: &product,
		PaymentRefer: stock.PaymentRefer,
		Payment: &payment,
		CreatedAt: stock.Product.CreatedAt,
		Note: stock.Note,
		Change: stock.Change,
		OldStock: stock.OldStock,
		NewStock: stock.NewStock,
		Type: stock.Type,
	}
}
func CreateNewNote (c *fiber.Ctx) error {
	type ParamsCreate struct {
		ProductRefer int `json:"product_id" validate:"required"`
		Product 	*Product `json:"product"`
		PaymentRefer int 	`json:"payment_id"`
		Payment 	*Payment	`gorm:"foreignKey:PaymentRefer"`
		Note 		string 	`json:"note"`
		Change 		int32 	`json:"change"`
		OldStock	int32 	`json:"old_stock"`
		NewStock 	int32 	`json:"new_stock"`
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
	if err := database.Database.DB.Find(&payment, "id = ?",params.PaymentRefer).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	if err := database.Database.DB.Where("status = ?").First(&payment).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	var orders models.Order
	if err := FindIdOrder(int(payment.OrderID), &orders); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	var products models.Product
	if err := findID(int(params.ProductRefer), &products); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id product!")
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
		ProductRefer: params.ProductRefer,
		Product: &products,
		PaymentRefer: params.PaymentRefer,
		Payment: &payment,
		Note: note,
		Change: int32(changeFinal),
		OldStock: params.OldStock,
		NewStock: params.NewStock,
		Type: "Order",
	}

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&result).Error; err != nil {
		tx.Rollback()
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	if err := tx.Save(&products).Error; err != nil {
		tx.Rollback()
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseStockLog := DataBaseIntoStock(stock, Product(products), DatabaseIntoPayment(payment))
	return utils.JsonWithSuccess(c, responseStockLog, fiber.StatusOK, "Berhasil membuat catatan order!")
}



