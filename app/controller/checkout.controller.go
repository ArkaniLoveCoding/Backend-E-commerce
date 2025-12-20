package controller

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ArkaniLoveCoding/fiber-project/database"
	"github.com/ArkaniLoveCoding/fiber-project/models"
	"github.com/ArkaniLoveCoding/fiber-project/utils"
)

type Checkout struct {
	ID 			uint `json:"id" gorm:"primaryKey"`
	CreatedAt 	time.Time
	OrderRefer   int   `json:"order_id" validate:"required"`
	Order		*Order  `json:"order"`
	Nominal 	float64  `json:"nominal" validate:"required"`
	Status 		string 	 `json:"status"`
}

func ResponeToCheckout (checkout models.Checkout, order Order) Checkout {
	return Checkout{
		ID: checkout.ID,
		CreatedAt: checkout.CreatedAt,
		OrderRefer: checkout.OrderRefer,
		Order: &order,
		Nominal: checkout.Nominal,
		Status: checkout.Status,
	}
}

func CreateCheckout (c *fiber.Ctx) error {
	type CreateParams struct {
		ID 			uint `json:"id" gorm:"primaryKey"`
		CreatedAt 	time.Time
		OrderRefer   int   `json:"order_id" validate:"required"`
		Order		*Order  `json:"order"`
		Nominal 	float64 `json:"nominal"`
	}
	var checkoutParams CreateParams
	if err := c.BodyParser(&checkoutParams); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal melakukan konfirmasi data!")
	}
	
	var order models.Order
	if err := database.Database.DB.
	Select("id").
	Preload("Product").
	Preload("User").Find(&order, "id = ?", checkoutParams.OrderRefer).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	// pengecekan validator
	var validate *validator.Validate = validator.New()
	if err := validate.Struct(checkoutParams); err != nil {
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


	paramsFinal := models.Checkout{
		OrderRefer: checkoutParams.OrderRefer,
		Nominal: order.TotalOrder,
		Status: "pending",
	}

	tx := database.Database.DB.Begin()
	defer func () {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if order.Product.Stock == 0 {
		tx.Rollback()
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Stock nya habis!")
	}

	if err := tx.Create(&paramsFinal).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal membuat suatu data!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseProduct := CreateProductResponse(*order.Product)
	responseUser := CreateUserResponse(*order.User)
	responseOrder := ResponseToOrder(order, responseUser, responseProduct)

	responseCheckout := ResponeToCheckout(paramsFinal, responseOrder)
	
	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseCheckout, fiber.StatusOK, "Berhasil membuat data checkout!")
}
func GetAllCheckout (c *fiber.Ctx) error {
	var order models.Order
	checkout := []models.Checkout{}
	Checkout := []Checkout{}

	sort := c.Query("sort")
	dir := c.Query("dir")

	ValidSort := map[string]bool {
		"nominal": true,
		"created_at": true,
	}

	ValidDir := map[string]bool {
		"asc": true,
		"desc": true,
	}

	if !ValidSort[sort] {
		sort = "created_at"
	}

	if !ValidDir[dir] {
		dir = "asc"
	}

	if err := database.Database.DB.
	Select("id, order_id, nominal, status").
	Scopes(database.Pagination(c)).
	Find(&checkout).Error; err != nil {
		if errors.Is(err, gorm.ErrInvalidData) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mendapatkan data!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mendapatkan semua data checkout!")
	}

	for _, result := range checkout{
		responseProduct := CreateProductResponse(*order.Product)
		responseUser := CreateUserResponse(*order.User)
		responseOrder := ResponseToOrder(order, responseUser, responseProduct)
		responseCheckout := ResponeToCheckout(result, responseOrder)
		Checkout = append(Checkout, responseCheckout)
	}
	return utils.JsonWithSuccess(c, Checkout, fiber.StatusOK, "Berhasil mengambil semua data checkout!")
}
func FindIdCheckout (id int, checkout *models.Checkout) error {
	if err := database.Database.DB.
	Select("id").
	Find(&checkout, "id = ?", id); err != nil {
		return nil
	}
	if checkout.ID == 0 {
		return nil
	}
	return nil
}
func GetOneCheckout (c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id!")
	}

	var order models.Order
	var checkout models.Checkout
	if err := FindIdCheckout(id, &checkout); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id untuk kedua kalinya!")
	}

	if err := database.Database.DB.First(&checkout).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan data yang diinginkan!")
	}

	responseUser := CreateUserResponse(*order.User)
	responseProduct := CreateProductResponse(*order.Product)
	responseOrder := ResponseToOrder(order, responseUser, responseProduct)
	responseCheckout := ResponeToCheckout(checkout, responseOrder)

	return utils.JsonWithSuccess(c, responseCheckout, fiber.StatusOK, "Berhasil mengambil salah satud data checkout!")
}
func DeleteCheckout (c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id!")
	}

	var order models.Order
	var checkout models.Checkout
	if err := FindIdCheckout(id, &checkout); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id untuk yang kedua kalinya!")
	}

	tx := database.Database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Delete(&checkout).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidData) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menghapus data!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menghapus data checkout!")
	}
	
	responseUser := CreateUserResponse(*order.User)
	responseProduct := CreateProductResponse(*order.Product)
	responseOrder := ResponseToOrder(order, responseUser, responseProduct)
	responseCheckout := ResponeToCheckout(checkout, responseOrder)
	
	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseCheckout, fiber.StatusOK, "Berhasil menghapus data yang diinginkan!")
}
func UpdateCheckout (c *fiber.Ctx) error {
	type UpdateParams struct {
		ID 			uint `json:"id" gorm:"primaryKey"`
		CreatedAt 	time.Time
		OrderRefer   int   `json:"order_id" validate:"required"`
		Order		Order  `json:"order"`
		Nominal 	float64  `json:"nominal" validate:"required"`
		Status 		string 	 `json:"status"`
	}

	var update UpdateParams
	if err := c.BodyParser(&update); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menkonfirmasi type data!")
	}

	// pengecekan validator
	var validate *validator.Validate = validator.New()
	if err := validate.Struct(update); err != nil {
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

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id!")
	}

	var checkout models.Checkout
	if err := FindIdCheckout(id, &checkout); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id untuk yang kedua kalinya")
	}

	var order models.Order
	if err := database.Database.DB.Preload("Product").Preload("User").Find(&order, "id = ?", checkout.OrderRefer).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id order!")
	}

	checkout.ID = update.ID
	checkout.CreatedAt = update.CreatedAt
	checkout.OrderRefer = update.OrderRefer
	checkout.Order = &order
	checkout.Nominal = update.Nominal

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Save(&checkout).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengubah data!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengubah data !")
	}

	responseUser := CreateUserResponse(*order.User)
	responseProduct := CreateProductResponse(*order.Product)
	responseOrder := ResponseToOrder(order, responseUser, responseProduct)
	responseCheckout := ResponeToCheckout(checkout, responseOrder)
	
	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseCheckout, fiber.StatusOK, "Berhasil mengubah data checkout!")
}
func PatchNominal (c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id!")
	}

	var checkout models.Checkout
	if err := FindIdCheckout(id, &checkout); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id untuk untuk kedua kalinya")
	}

	var order models.Order
	if err := database.Database.DB.
	Select("id").
	Preload("Product").
	Preload("User").Find(&order, "id = ?", checkout.OrderRefer).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal melihat order!")
	}

	inputsNominal := make(map[string]interface{})
	if err := c.BodyParser(&inputsNominal); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengkonvert data!")
	}

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()


	if err := tx.
	Select("nominal").
	Model(&checkout).
	Updates(inputsNominal).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengubah data!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseUser := CreateUserResponse(*order.User)
	responseProduct := CreateProductResponse(*order.Product)
	responseOrder := ResponseToOrder(order, responseUser, responseProduct)
	responseCheckout := ResponeToCheckout(checkout, responseOrder)
	
	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseCheckout, fiber.StatusOK, "Berhasil mengubah data checkout!")
}
func BatchAllDeleteCheckout (c *fiber.Ctx) error {
	id  := c.Query("id")

	var paramsId []uint
	idStr := strings.Split(id, " ")
	for _, resultId := range idStr {
		idInt, err := strconv.Atoi(resultId)
		if err != nil {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengolah id!")
		}
		paramsId = append(paramsId, uint(idInt))
	}

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var checkout models.Checkout
	if err := tx.
	Select("id").
	Where("id ? IN", paramsId).
	Delete(&checkout).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menghapus data!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengolah data!")
	}

	var order models.Order
	responseUser := CreateUserResponse(*order.User)
	responseProduct := CreateProductResponse(*order.Product)
	responseOrder := ResponseToOrder(order, responseUser, responseProduct)
	responseCheckout := ResponeToCheckout(checkout, responseOrder)

	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseCheckout, fiber.StatusOK, "Berhasil menghapus dan mengupdate data sekaligus!")
}

// statistik perusahaan 
func AvarageExpenseUser (c *fiber.Ctx) error {
	type AvarageParams struct {
		UserName 	string  `json:"user_name" gorm:"column:user_name"`
		TotalAvarage 	string `json:"total_avarage" gorm:"column:total_avarage"`
	}

	var ResultTotal []AvarageParams 
	if err := database.Database.DB.Table("checkouts").
	Select("users.name AS user_name, AVG(orders.total_order * checkouts.nominal) AS total_avarage").
	Joins("JOIN orders ON checkout.order_refer = orders.id").
	Group("users.name"). 
	Scan(&ResultTotal).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, ResultTotal, fiber.StatusOK, "Berhasil mendapaatkan data!")
}
func TotalCheckouts (c *fiber.Ctx) error {
	type Params struct {
		OrderID 	uint 	`json:"order_id" gorm:"column:order_id"`
		Name 		string 	`json:"user_name" gorm:"column:user_name"`
		Product 	string	`json:"product_name" gorm:"column:product_name"`
		TotalCheckout int 	`json:"total_checkout" gorm:"column:total_checkout"`
	}

	var TotalResult []Params
	if err := database.Database.DB.Table("checkouts"). 
	Select("checkouts.order_refer AS order_id, users.name AS user_name, SUM(checkouts.id) AS total_checkouts, products.name AS product_name"). 
	Joins("JOIN orders ON checkouts.order_refer = orders.id"). 
	Joins("JOIN users ON orders.user_refer = users.id"). 
	Joins("JOIN products ON orders.product_refer = products.id"). 
	Group("checkouts.order_refer, products.name, users.name").Order("total_checkouts desc").
	Scan(&TotalResult).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, TotalResult, fiber.StatusOK, "Berhasil mendapatkan data!")
}

