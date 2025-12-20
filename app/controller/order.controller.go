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

type Order struct {
	ID 			uint `json:"id" gorm:"primaryKey"`
	CreatedAt 	time.Time
	Order 		time.Time
	Quantity  	int32 `json:"quantity"`
	Status 		string `json:"status"`
	TotalOrder 	float64 `json:"total_order"`
	ProductRefer int `json:"product_id"`
	Product		*Product  `json:"product"`
	UserRefer 	int   `json:"user_id"`
	User 		*User	`json:"user"`
}
func ResponseToOrder (order models.Order, user User, product Product) Order {
	return Order{
		ID: order.ID,
		CreatedAt: order.CreatedAt,
		Order: order.Order,
		Quantity: order.Quantity,
		Status: order.Status,
		TotalOrder: order.TotalOrder,
		ProductRefer: order.ProductRefer,
		Product: &product,
		UserRefer: order.UserRefer,
		User: &user,
	}
}

func CreateOrder (c *fiber.Ctx) error {
	type ParamsForCreate struct {
		ID 				uint 	`json:"id" gorm:"primaryKey"`
		CreatedAt 		time.Time
		Order 			time.Time
		Quantity  		int32 `json:"quantity" validate:"required"`
		Status 			string `json:"status"`
		TotalOrder 		float64 `json:"total_order"`
		ProductRefer 	int 	`json:"product_id" validate:"required"`
		Product 		Product `gorm:"foreignKey:ProductRefer"`
		UserRefer 		int 	`json:"user_id" validate:"required"`
		User 			User 	`gorm:"foreignKey:UserRefer"`
	}
	var orders ParamsForCreate
	if err := c.BodyParser(&orders); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memvalidasi body!")
	}

		// pengecekan validator
	var validate *validator.Validate = validator.New()
	if err := validate.Struct(orders); err != nil {
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

	var product models.Product
	if err := findID(orders.ProductRefer, &product); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memuat id product!")
	}

	var user models.User
	if err := FindIdUser(orders.UserRefer, &user); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memuat id user!")
	}


	// hitung total order yang dibuat oleh user
	orders.TotalOrder = float64(orders.Quantity) * product.Price

	result := models.Order{
		ID: orders.ID,
		CreatedAt: orders.CreatedAt,
		Quantity: orders.Quantity,
		Status: "pending",
		TotalOrder: orders.TotalOrder,
		Order: orders.Order,
		ProductRefer: orders.ProductRefer,
		Product: &product,
		UserRefer: orders.UserRefer,
		User: &user,
	}

	tx := database.Database.DB.Begin()
	defer func () {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// kita memakai ini agar bisa mengambil data hasil dari query yang kita masukkan
	if err := tx.Create(&result).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal membuat order baru!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseProduct := CreateProductResponse(product)
	responseUser := CreateUserResponse(user)
	responseCreate := ResponseToOrder(result, responseUser, responseProduct)
	
	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseCreate, fiber.StatusOK, "Berhasil menambahkan order!")
}
func GetAllOrder (c *fiber.Ctx) error {
	orders := []models.Order{}
	Order := []Order{}

	sort := c.Query("sort")
	dir := c.Query("dir")

	ValidSort := map[string]bool {
		"total_order": true,
		"created_at": true,
	}

	if !ValidSort[sort] {
		sort = "created_at"
	}

	ValidDir := map[string]bool {
		"asc": true,
		"desc": true,
	}

	if !ValidDir[dir] {
		dir = "asc"
	}

	if err := database.Database.DB.
	Select("id, quantity, status, total_order, product_id, user_id").
	Scopes(database.
	Pagination(c)).
	Find(&orders).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}
	if len(orders) == 0 {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Tidak ada data yang terlampir!")
	}
	
	var product models.Product
	var user models.User
	for _, result := range orders {
		responseProduct := CreateProductResponse(product)
		responseUser := CreateUserResponse(user)
		responseOrder := ResponseToOrder(result, responseUser, responseProduct)
		Order = append(Order, responseOrder)
	}
	return utils.JsonWithSuccess(c, Order, fiber.StatusOK, "Berhasil melihat semua data order!")
}
func FindIdOrder (id int, order *models.Order) error {
	if err := database.Database.DB.
	Select("id").
	Find(&order, "id = ?", id).Error; err != nil {
		return errors.New(err.Error())
	}
	if order.ID == 0 {
		return nil
	}
	return nil
}

func GetOneOrder (c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memvalidasi id!")
	}

	var order models.Order
	if err := FindIdOrder(id, &order); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memvalidasi pencarian id!")
	}

	var product models.Product
	if err := findID(order.ProductRefer, &product); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mendapatkan id product!")
	}

	var user models.User
	if err := FindIdUser(order.UserRefer, &user); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mendapatkan id user!")
	}
	// kita memakai ini agar bisa mengambil data hasil dari query yang kita masukkan
	responseProduct := CreateProductResponse(product)
	responseUser := CreateUserResponse(user)
	responseOrder := ResponseToOrder(order, responseUser, responseProduct)
	return utils.JsonWithSuccess(c, responseOrder, fiber.StatusOK, "Berhasil mengambil data salah satu order!")
}
func DeleteOrder (c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mencari id!")
	}

	var order models.Order
	if err := FindIdOrder(id, &order); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memvalidasi id!")
	}

	var product models.Product
	if err := findID(order.ProductRefer, &product); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id product!")
	}

	var user models.User
	if err := FindIdUser(order.UserRefer, &user); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id user!")
	}

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// kita memakai ini agar bisa mengambil data hasil dari query yang kita masukkan
	if err := tx.Delete(&order).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menghapus data!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseProduct := CreateProductResponse(product)
	responseUser := CreateUserResponse(user)
	responseOrder := ResponseToOrder(order, responseUser, responseProduct)
	
	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseOrder, fiber.StatusOK, "Berhasil menghapus data order!")
}
func UpdateOrder (c *fiber.Ctx) error {
	type ParamsForUpdate struct {
		ID 				uint 	`json:"id" gorm:"primaryKey"`
		CreatedAt 		time.Time
		Order 			time.Time
		Quantity 		int32 	`json:"quantity" validate:"required"`
		Status 			string 	`json:"status"`
		TotalOrder 		float64  `json:"total_order"`
		ProductRefer 	int 	`json:"product_id" validate:"required"`
		Product 		Product `gorm:"foreignKey:ProductRefer"`
		UserRefer 		int 	`json:"user_id" validate:"required"`
		User 			User 	`gorm:"foreignKey:UserRefer"`
	}

	var paramsUpdate ParamsForUpdate
	if err := c.BodyParser(&paramsUpdate); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal melakukan memvalidasi body!")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mencari id!")
	}

	var order models.Order
	if err := FindIdOrder(id, &order); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memvalidasi id!")
	}

	var user models.User
	if err := FindIdUser(paramsUpdate.UserRefer, &user); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengambil id user!")
	}

	var product models.Product
	if err := findID(paramsUpdate.ProductRefer, &product); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan atau mengambil id product!")
	}

	// pengecekan validator
	var validate *validator.Validate = validator.New()
	if err := validate.Struct(paramsUpdate); err != nil {
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

	paramsUpdate.TotalOrder = float64(paramsUpdate.Quantity) * product.Price

	order.ProductRefer = paramsUpdate.ProductRefer 
	order.UserRefer = paramsUpdate.UserRefer
	order.Quantity = paramsUpdate.Quantity
	order.Status = paramsUpdate.Status
	order.TotalOrder = paramsUpdate.TotalOrder

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengubah data!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseProduct := CreateProductResponse(product)
	responseUser := CreateUserResponse(user)
	responseOrder := ResponseToOrder(order, responseUser, responseProduct)
	
	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseOrder, fiber.StatusOK, "Berhasil mengupdate data order!")
}
func PatchQuantityOrder (c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id!")
	}
	
	inputsUser := make(map[string]interface{})
	if err := c.BodyParser(&inputsUser); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal membuat fields!")
	}

	var order models.Order
	if err := FindIdOrder(id, &order); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id kembali!")
	}

	var product models.Product
	if err := database.Database.DB.
	Select("id").
	First(&product, order.ProductRefer).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}
	if quantity, ok := inputsUser["quantity"]; ok {
		qty := int32(quantity.(float64))
		
		total := float64(qty) * product.Price

		inputsUser["total_order"] = total
	}

	var user models.User
	if err := FindIdUser(order.UserRefer, &user); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengubah quantity!")
	}	
	
	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.
	Select("id, quantity, total_order, status, product_id, user_id").
	Model(&order).
	Updates(inputsUser).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengubah data!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseUser := CreateUserResponse(user)
	responseProduct := CreateProductResponse(product)
	responseOrder := ResponseToOrder(order, responseUser, responseProduct)
	
	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseOrder, fiber.StatusOK, "Berhasil menambahkan data!")
}
func SearchProductAndUser (c *fiber.Ctx) error {
	keyword := c.Query("search")

	var orders []models.Order

	if err := database.Database.DB.
	Where("CAST(product_refer AS TEXT) LIKE ?", "%"+keyword+"%").
	Preload("User").
	Preload("Product").
	Find(&orders).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	result := make([]Order, len(orders))
	for x, i := range orders {
		responseUser := CreateUserResponse(*i.User)
		responseProduct := CreateProductResponse(*i.Product)
		result[x] = ResponseToOrder(i, responseUser, responseProduct)
	}
	return utils.JsonWithSuccess(c, result, fiber.StatusOK, "Berhasil mengambil hasil pencarian order!")
}
func BatchAllDeleteOrder (c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengolah kueri!")
	}

	var paramsId []uint
	idStr := strings.Split(id, " ")
	for _, resultId := range idStr {
		idInt, err := strconv.Atoi(resultId)
		if err != nil {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengolah id!")
		}
		paramsId = append(paramsId, uint(idInt))
	}
	
	var order models.Order
	var product models.Product
	if err := findID(order.ProductRefer, &product); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	var user models.User
	if err := FindIdUser(order.UserRefer, &user); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.
	Select("id").
	Preload("Product").
	Preload("User").Where("id ? IN", paramsId).Delete(&order).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menghapus data!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseUser := CreateUserResponse(user)
	responseProduct := CreateProductResponse(product)
	responseOrder := ResponseToOrder(order, responseUser, responseProduct)

	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseOrder, fiber.StatusOK, "Berhasil menghapus data sekaligus!")
}
func BatchAllUpdateOrder (c *fiber.Ctx) error {
	type ParamsForUpdate struct {
		ID 				uint 	`json:"id" gorm:"primaryKey"`
		CreatedAt 		time.Time
		Order 			time.Time
		Quantity 		int32 	`json:"quantity" validate:"required"`
		TotalOrder 		float64  `json:"total_order"`
		ProductRefer 	int 	`json:"product_id" validate:"required"`
		Product 		Product `gorm:"foreignKey:ProductRefer"`
		UserRefer 		int 	`json:"user_id" validate:"required"`
		User 			User 	`gorm:"foreignKey:UserRefer"`
	}
	var updateOrder ParamsForUpdate
	if err := c.BodyParser(&updateOrder); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengolah data!")
	}

	id := c.Query("id")

	idStr := strings.Split(id, " ")
	var idUint []uint

	for _, result := range idStr {
		idInt, err := strconv.Atoi(result)
		if err != nil {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengubah data ke int!")
		}
		idUint = append(idUint, uint(idInt))
	}

	var order models.Order
	order.ProductRefer = updateOrder.ProductRefer 
	order.UserRefer = updateOrder.UserRefer
	order.Quantity = updateOrder.Quantity
	order.TotalOrder = updateOrder.TotalOrder

	var product models.Product
	if err := findID(order.ProductRefer, &product); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id product!")
	}

	var user models.User
	if err := FindIdUser(order.UserRefer, &user); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id user!")
	}

	tx := database.Database.DB.Begin()
	defer func (){ 
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()
	if err := tx.
	Select("id").
	Model(&order).
	Where("id ? IN", idUint).
	Updates(updateOrder).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseUser := CreateUserResponse(user)
	responseProduct := CreateProductResponse(product)
	responseOrder := ResponseToOrder(order, responseUser, responseProduct)

	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseOrder, fiber.StatusOK, "Berhasil mengubah semua data yang diinginkan!")
}

// bagian statistik perusahaan (best practice)
func TotalOrderForManyUsers (c *fiber.Ctx) error {
	type TotalOrderManyUsers struct {
		UserId uint  `json:"user_id" gorm:"column:user_id"`
		Name string `json:"user_name" gorm:"column:user_name"`
		TotalOrder float64  `json:"total_orders" gorm:"column:total_orders"`
	}

	var result []TotalOrderManyUsers
	if err := database.Database.DB.Table("orders").
	Select("orders.user_refer AS user_id, SUM(products.price * orders.quantity) AS total_orders, users.name AS user_name").
	Joins("JOIN users ON orders.user_refer = users.id"). 
	Joins("JOIN products ON orders.product_refer = products.id").
	Group("orders.user_refer, users.name").Scan(&result).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, result, fiber.StatusOK, "Berhasil mendapatkan data!")
}
func TotalQtyProduct (c *fiber.Ctx) error {
	type Total struct {
		ProductID int `json:"product_id" gorm:"column:product_id"`
		TotalQty int `json:"total_qty" gorm:"column:total_qty"`
		Name string `json:"user_name" gorm:"column:user_name"`
	}

	var TotalResult []Total
	if err := database.Database.DB.Table("orders").
	Select("orders.product_refer AS product_id, SUM(orders.quantity) AS total_qty, users.name AS user_name").
	Joins("JOIN products ON orders.product_refer = products.id").
	Joins("JOIN users ON orders.user_refer = users.id").
	Group("orders.product_refer, users.name").Order("total_qty desc").
	Scan(&TotalResult).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, TotalResult, fiber.StatusOK, "Berhasil mendapatkan data!")
}
func ProductsStatistikToOrder (c *fiber.Ctx) error {
	type ProductStatistik struct {
		ProductID 	uint 	`json:"product_id" gorm:"column:product_id"`
		Name 		string `json:"user_name" gorm:"column:user_name"`
		TotalQty 	int 	`json:"total_qty" gorm:"column:total_qty"`
		TotalIncome float64 `json:"total_income" gorm:"column:total_income"`
		OrderQty 	int 	`json:"order_qty" gorm:"column:order_qty"`
	}

	var ProductStat []ProductStatistik
	if err := database.Database.DB.Table("orders").
	Select(`orders.product_refer AS product_id, 
	SUM(products.id) AS total_qty, SUM(orders.id * products.price) AS total_income, 
	SUM(orders.id) AS order_qty, users.name AS user_name`).
	Joins("JOIN products ON orders.product_refer = products.id").
	Joins("JOIN users ON orders.user_refer = users.id").
	Group("orders.product_refer, users.name").Order("total_qty").
	Scan(&ProductStat).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, ProductStat, fiber.StatusOK, "Berhasil mendapatkan data statistik product!")
}
func CountAndSumOrderAndQuantity (c *fiber.Ctx) error {
	type ParamsTotal struct {
		TotalOrder float64 `json:"total_order" gorm:"column:total_order"`
		TotalQuantity int `json:"total_quantity" gorm:"column:total_quantity"`
	}

	var TotalResult []ParamsTotal
	if err := database.Database.DB.Table("orders").
	Select("SUM(orders.quantity) AS total_quantity, SUM(orders.total_order) AS total_order").
	Scan(&TotalResult).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, TotalResult, fiber.StatusOK, "Berhasil mendapatkan data!")
}
func FindOrderUserToProduct (c *fiber.Ctx) error {
	type Params struct {
		ProductID uint `json:"product_id" gorm:"column:product_id"`
		ProductName string `json:"product_name" gorm:"column:product_name"`
		ProductPrice float64 `json:"product_price" gorm:"column:product_price"`
		OrderID   uint  `json:"order_id" gorm:"column:order_id"`
		Name 	  string `json:"user_name" gorm:"column:user_name"`
		TotalQty  int 	  `json:"total_quantity" gorm:"column:total_quantity"`
	}

	var Total []Params
	if err := database.Database.DB.Table("orders"). 
	Select("orders.id AS order_id, SUM(orders.quantity) AS total_quantity, users.name AS user_name, orders.product_refer AS product_id, products.name AS product_name, products.price AS product_price"). 
	Joins("JOIN users ON orders.user_refer = users.id"). 
	Joins("JOIN products ON orders.product_refer = products.id"). 
	Group("orders.product_refer, users.name, orders.id, products.name, products.price").
	Order("total_quantity desc").Scan(&Total).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, Total, fiber.StatusOK, "Berhasil mendapatkan data!")
}
func FindUserOrder (c *fiber.Ctx) error {
	type Params struct {
		UserID 		uint 	`json:"user_id" gorm:"column:user_id"`
		UserName 	string	`json:"user_name" gorm:"column:user_name"`
		ProductName string	`json:"product_name" gorm:"product_name"`
		OrderID 	uint 	`json:"order_id" gorm:"order_id"`
	}

	var Result []Params
	if err := database.Database.DB.Table("orders"). 
	Select("orders.user_refer AS user_id, products.name AS product_name, users.name AS user_name, orders.id AS order_id"). 
	Joins("JOIN products ON orders.product_refer = products.id"). 
	Joins("JOIN users ON orders.user_refer = users.id"). 
	Group("users.name, orders.user_refer, products.name"). 
	Scan(&Result).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, Result, fiber.StatusOK, "Berhasil mendapatkan data!")
}
func FindOrderWithId (c *fiber.Ctx) error {
	params := c.Query("product_name")

	type Params struct {
		UserName 	string 	`json:"user_name" gorm:"column:user_name"`
		ProductName string 	`json:"product_name" gorm:"column:product_name"`
		OrderID 	uint 	`json:"order_id" gorm:"column:order_id"`
		TotalQuantity 	int 	`json:"total_quantity" gorm:"column:total_quantity"`
	}

	var Result []Params
	if err := database.Database.DB.Table("orders"). 
	Select("orders.id AS order_id, users.name AS user_name, SUM(orders.quantity) AS total_quantity, products.name AS product_name").
	Joins("JOIN users ON orders.user_refer = users.id"). 
	Joins("JOIN products ON orders.product_refer = products.id").
	Group("users.name, orders.id, products.name").
	Where("products.name = ?", params).
	Order("total_quantity desc").Scan(&Result).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	} 

	return utils.JsonWithSuccess(c, Result, fiber.StatusOK, "Berhasil mendapatkan data!")
}
func TotalQtyFromLow (c *fiber.Ctx) error {
	type Total struct {
		ProductID int `json:"product_id" gorm:"column:product_id"`
		TotalQty int `json:"total_qty" gorm:"column:total_qty"`
		Name string `json:"user_name" gorm:"column:user_name"`
	}

	var TotalResult []Total
	if err := database.Database.DB.Table("orders").
	Select("orders.product_refer AS product_id, SUM(orders.quantity) AS total_qty, users.name AS user_name").
	Joins("JOIN products ON orders.product_refer = products.id").
	Joins("JOIN users ON orders.user_refer = users.id").
	Group("orders.product_refer, users.name").Order("total_qty asc").
	Scan(&TotalResult).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, TotalResult, fiber.StatusOK, "Berhasil mendapatkan data!")
}