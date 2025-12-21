package controller

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
type Product struct {
	ID 			uint `json:"id"`
	CreatedAt 	time.Time
	Name		string `json:"name"`
	Price 		float64 `json:"price"`
	Stock 		int32   `json:"stock"`
	Serialnumber string `json:"serial_number"`
	Expired 	string  `json:"expired"`
	Category 	string  `json:"category"`
	Image 		string  `json:"image"`
	Status 		string  `json:"status"`
}
func CreateProductResponse (product models.Product) Product {
	return Product{
		ID: product.ID,
		CreatedAt: product.CreatedAt,
		Name: product.Name,
		Price: product.Price,
		Stock: product.Stock,
		Serialnumber: product.Serialnumber,
		Expired: product.Expired,
		Category: product.Category,
		Image: product.Image,
		Status: product.Status,
	}
}
func CreatenNewProduct (c *fiber.Ctx) error {
	// parse ke form value go fiber
	nameValue := c.FormValue("name")
	priceValue := c.FormValue("price")
	stockValue := c.FormValue("stock")
	serialNumberValue := c.FormValue("serial_number")
	expiredValue := c.FormValue("expired")
	categoryValue := c.FormValue("category")
	statusValue := c.FormValue("status")


	stockConvert, err := strconv.Atoi(stockValue)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengkonvert ke int32!")
	}
	priceConvert, err := strconv.ParseFloat(priceValue, 64)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengconvert ke float64!")
	}
	stockFinal := int32(stockConvert)
	priceFinal := float64(priceConvert)


	// handler file path to image 
	file, err := c.FormFile("image")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	// buat nama parameter tempat save file
	uploadDirName := "/uploads"
	if _, err := os.Stat(uploadDirName); os.IsNotExist(err) {
		os.MkdirAll(uploadDirName, os.ModePerm)
	}

	// pengecekan file
	fileName := file.Filename
	filePathFinal := filepath.Join(uploadDirName, fileName)
	
	if file.Header.Get("Content-Type") != "image/png" && 
	file.Header.Get("Content-Type") != "image/jpg" && 
	file.Header.Get("Content-Type") != "image/jpeg" {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Type yang dimasukkan salah!")
	}

	paramsProduct := models.Product{
		Name: nameValue,
		Price: priceFinal,
		Stock: stockFinal,
		Serialnumber: serialNumberValue,
		Expired: expiredValue,
		Category: categoryValue,
		Image: filePathFinal,
		Status: statusValue,
	}

	// pengecekan validator
	var validate *validator.Validate = validator.New()
	if err := validate.Struct(paramsProduct); err != nil {
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

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Create(&paramsProduct).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal membuat product!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}
	responseProduct := CreateProductResponse(paramsProduct)
	
	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseProduct, fiber.StatusOK, "Berhasil membuat product!")
}
func GetAllProducts (c *fiber.Ctx) error {
	products := []models.Product{}
	responseProducts := []Product{}

	sort := c.Query("sort")
	dir := c.Query("dir")

	ValidSort := map[string]bool {
		"price": true,
		"name": true,
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
	Select("id, name, stock, price, category, status, expired, serial_number, image").
	Scopes(database.Pagination(c)).
	Order(fmt.Sprintf("%s %s", sort, dir)).
	Find(&products).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}
	for _, products := range products {
		responseProduct := CreateProductResponse(products)
		responseProducts = append(responseProducts, responseProduct)
	}
	return utils.JsonWithSuccess(c, products, fiber.StatusOK, "Berhasil mendapatkan semua data produk!")
}
func findID (id int, product *models.Product) error {
	if err := database.Database.DB.
	Select("id", "name", "price", "stock", "expired", "serial_number", "category", "image", "status").
	Where("id = ?", id).
	Find(&product).Error; err != nil {
		return errors.New(err.Error())
	}
	if product.ID == 0 {
		return errors.New("Gagal mendapatkan id!")
	}
	return nil
}
func GetOneProduct (c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Id hanya boleh bertipe integer")
	}

	var products models.Product
	if err := findID(id, &products); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengambil id yang diminta!")
	}
	
	responseProduct := CreateProductResponse(products)

	return utils.JsonWithSuccess(c, responseProduct, fiber.StatusOK, "Berhasil mengambil salah satu data menggunakan id!")
}
func UpdateProduct (c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal dalam mengammbil id!")
	}

	var products models.Product
	if err := findID(id, &products); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengambil id dari user!")
	}

	// parsing data ke form value 
	namaValue := c.FormValue("name")
	priceValue := c.FormValue("price")
	stockValue := c.FormValue("stock")
	serialNumberValue := c.FormValue("serial_number")
	expiredValue := c.FormValue("expired")
	categoryValue := c.FormValue("category")
	statusValue := c.FormValue("status")


	priceConvert, err := strconv.ParseFloat(priceValue, 64)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengconvert ke float!")
	}
	stockConvert, err := strconv.Atoi(stockValue)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengconvert data!")
	}
	stockFinal := int32(stockConvert)
	priceFinal := float64(priceConvert)

	// mengatur path yang sesuai dengan request client
	file, err := c.FormFile("image")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	uploadDir := "/uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, os.ModePerm)
	}

	fileName := file.Filename
	filePath := filepath.Join(uploadDir, fileName)

	if file.Header.Get("Content-Type") != "image/png" && 
	file.Header.Get("Content-Type") != "image/jpg" &&
	file.Header.Get("Content-Type") != "image/jpeg" {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Format gambar tidak valid!")
	}

	
	products.Name = namaValue
	products.Price = priceFinal
	products.Stock = stockFinal
	products.Serialnumber = serialNumberValue
	products.Expired = expiredValue
	products.Category = categoryValue
	products.Image = filePath
	products.Status = statusValue

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Save(&products).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengupdate data!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}
	
	responseProduct := CreateProductResponse(products)
	
	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseProduct, fiber.StatusOK, "Berhasil mengupdate data product!")
}
func PatchProduct (c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mendapatkan id product!")
	}

	var product models.Product
	if err := findID(id, &product); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan id yang diinginkan!")
	}

	inputsResult := make(map[string]interface{})
	if err := c.BodyParser(&inputsResult); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	price := c.FormValue("price")
	stockStr := c.FormValue("stock")

	if stockStr != "" {
		result, err  := strconv.Atoi(stockStr)
		resultFinal := int32(result)
		if err != nil {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengkonvert data!")
		}
		inputsResult["stock"] = resultFinal
	}
	if price != "" {
		result, err := strconv.ParseFloat(price, 32)
		resultFinal := float64(result)
		if err != nil {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengkonvert data!")
		}
		inputsResult["price"] = resultFinal
	}

	tx := database.Database.DB.Begin()
	defer func (){
	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	}
	}()

	if err := tx.
	Select("name, price, stock, serial_number, expired, category, image, status").
	Model(&product).
	Updates(inputsResult).Error; err != nil {
		tx.Rollback()
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}
	responseProduct := CreateProductResponse(product)
	
	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseProduct, fiber.StatusOK, "Berhasil mengupdate data!")
}
func DeleteProduct (c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengambil id!")
	}

	var products models.Product
	if err := findID(id, &products); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memvalidasi id!")
	}

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.Delete(products).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menghapus data!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menghapus data product!")
	}
	responseProduct := CreateProductResponse(products)
	
	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseProduct, fiber.StatusOK, "Berhasil menghapus data product!")
}
func SeacrhProductFromStatus (c *fiber.Ctx) error {
	keyword := c.Query("search")

	if keyword == "" {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Qeuery kosong!")
	}
	var products []models.Product
	
	if err := database.Database.DB.
	Select("status", "name", "price", "stock", "expired", "serial_number", "category", "image", "status").
	Where("status ILIKE ?", "%"+keyword+"%").
    Find(&products).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mendapatkan data!")
	}
	
	response := make([]Product, len(products))
	for i, p := range products {
		response[i] = CreateProductResponse(p)
	}
	return utils.JsonWithSuccess(c, response, fiber.StatusOK, "Berhasil mendapatkan data yang diinginkan!")
}
func SearchProductFromName (c *fiber.Ctx) error {
	paramsQuery := c.Query("search")

	if paramsQuery == "" {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Tidak ada query yang dimasukkan!")
	}

	var products []models.Product
	if err := database.Database.DB.
	Select("name", "price", "stock", "expired", "serial_number", "category", "image", "status").
	Where("name ILIKE ?", "%"+paramsQuery+"%").
    Find(&products).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mendapatkan data!")
	}

	result := make([]Product, len(products))
	for x, i := range products {
		result[x] = CreateProductResponse(i)
	}

	return utils.JsonWithSuccess(c, result, fiber.StatusOK, "Berhasil melihat data yang ingin dicari!")
}
func BatchAllDeleteProduct (c *fiber.Ctx) error {
	id := c.Query("id")

	var idUint []uint
	idStr := strings.Split(id, " ")
	for _, result := range idStr {
		idInt, err := strconv.Atoi(result)
		if err != nil {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal melakukan konvert ke type int!")
		}
		idUint = append(idUint, uint(idInt))
	}

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var product models.Product
	if err := tx.
	Select("id").
	Where("id ? IN", idUint).
	Delete(&product).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrInvalidField) {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menghapus data!")
		}
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseProduct := CreateProductResponse(product)
	
	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseProduct, fiber.StatusOK, "Berhasil menghapus data id !")
}
func BatchAllUpdateProduct (c *fiber.Ctx) error {
	// parsing data ke form value 
	namaValue := c.FormValue("name")
	priceValue := c.FormValue("price")
	stockValue := c.FormValue("stock")
	serialNumberValue := c.FormValue("serial_number")
	expiredValue := c.FormValue("expired")
	categoryValue := c.FormValue("category")
	statusValue := c.FormValue("status")


	priceConvert, err := strconv.ParseFloat(priceValue, 64)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengconvert ke float!")
	}
	stockConvert, err := strconv.Atoi(stockValue)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengconvert data!")
	}
	stockFinal := int32(stockConvert)
	priceFinal := float64(priceConvert)

	// mengatur path yang sesuai dengan request client
	file, err := c.FormFile("image")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	uploadDir := "/uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, os.ModePerm)
	}

	fileName := file.Filename
	filePath := filepath.Join(uploadDir, fileName)

	if file.Header.Get("Content-Type") != "image/png" && 
	file.Header.Get("Content-Type") != "image/jpg" &&
	file.Header.Get("Content-Type") != "image/jpeg" {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Format gambar tidak valid!")
	}

	var products models.Product
	products.Name = namaValue
	products.Price = priceFinal
	products.Stock = stockFinal
	products.Serialnumber = serialNumberValue
	products.Expired = expiredValue
	products.Category = categoryValue
	products.Image = filePath
	products.Status = statusValue

	id := c.Query("id")

	var idUint []uint
	idStr := strings.Split(id, " ")

	for _, result := range idStr {
		idInt, err := strconv.Atoi(result)
		if err != nil {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengkonversi ke type int!")
		}

		idUint = append(idUint, uint(idInt))
	}

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if err := tx.
	Select("name, price, stock, serial_number, expired, category, image, status").
	Model(&products).
	Where("id ? IN", idUint).
	Updates(products).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseProduct := CreateProductResponse(products)
	
	if err := tx.Commit().Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, responseProduct, fiber.StatusOK, "Berhasil mengupdate banyak data sekaligus!")
}
func GetProductPriceHighToLow (c *fiber.Ctx) error {
	type ParamsProduct struct {
		ProductID 	uint 	`json:"product_id" gorm:"column:product_id"`
		ProductName string 	`json:"product_name" gorm:"column:product_name"`
		ProductPrice float64 `json:"product_price" gorm:"column:product_price"`
		ProductCategory string `json:"product_category" gorm:"column:product_category"`
		ProductExpired string  `json:"product_expired" gorm:"column:product_expired"`
		ProductImage   string  `json:"product_image" gorm:"column:product_image"`
	}

	var ResultProduct []ParamsProduct
	if err := database.Database.DB.Table("products"). 
	Select("products.id AS product_id, products.name AS product_name, products.image AS product_image, products.category AS product_category, products.price AS product_price, products.expired AS product_expired"). 
	Group("products.id, products.name, products.image, products.category, products.expired, products.price"). 
	Order("products.price desc").Scan(&ResultProduct).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, ResultProduct, fiber.StatusOK, "Berhasil mendapatkan data!")
}
func GetProductStockHighToLow (c *fiber.Ctx) error {
	type ParamsProduct struct {
		ProductID 	uint 	`json:"product_id" gorm:"column:product_id"`
		ProductName string 	`json:"product_name" gorm:"column:product_name"`
		ProductStock int `json:"product_stock" gorm:"column:product_stock"`
		ProductCategory string `json:"product_category" gorm:"column:product_category"`
		ProductExpired string  `json:"product_expired" gorm:"column:product_expired"`
		ProductImage   string  `json:"product_image" gorm:"column:product_image"`
	}

	var ResultProduct []ParamsProduct
	if err := database.Database.DB.Table("products"). 
	Select("products.id AS product_id, products.name AS product_name, products.image AS product_image, products.category AS product_category, products.stock AS product_stock, products.expired AS product_expired"). 
	Group("products.id, products.name, products.image, products.category, products.expired, products.price"). 
	Order("products.stock desc").Scan(&ResultProduct).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JsonWithSuccess(c, ResultProduct, fiber.StatusOK, "Berhasil mendapatkan data!")
}