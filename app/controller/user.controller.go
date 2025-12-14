package controller

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"github.com/ArkaniLoveCoding/fiber-project/database"
	"github.com/ArkaniLoveCoding/fiber-project/models"
	"github.com/ArkaniLoveCoding/fiber-project/utils"
)


type User struct {
	ID 			uint `json:"id"`
	Name 		string `json:"name"`
	Password 	string `json:"password"`
	Email 		string `json:"email"`
	Role 		string `json:"role"`
}
func CreateUserResponse (user models.User) User {
	return User{
		ID: user.ID,
		Name: user.Name,
		Password: user.Password,
		Email: user.Email,
		Role: user.Role,
	}
} 
func CreateUserNew (c *fiber.Ctx) error {
	type UserRequest struct {
		ID 			uint `json:"id"`
		Name 		string `json:"name"`
		Password    string `json:"password" validate:"required,min=6"`
		Email  		string `json:"email" validate:"required,email"`
		Role 		string `json:"role" validate:"required"`
		AdminPassword string 	`json:"password_admin"`
	}
	var params UserRequest
	if err := c.BodyParser(&params) ; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memvalidasi body!")
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
	hash, err := utils.GenerateAndHashPassword(params.Password)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal melakukan hash password!")
	}

	if params.Role == "admin" {
		if err := godotenv.Load(); err != nil {
			return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
		}
		passwordAdmin := os.Getenv("AdminCode")
		if params.AdminPassword != passwordAdmin {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Pasword yang dimasukkan salah!, dan tidak ada!")
		}
		if params.AdminPassword == "" {
			return utils.JsonWithError(c, fiber.StatusBadRequest, "Jika memilih sebagai admin, harus memasukkan password adminb!")
		}
	}

	user := models.User{
		ID: params.ID,
		Name: params.Name,
		Password: hash,
		Email: params.Email,
		Role: params.Role,
	}


	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	// json web token into generate at here
	jwt, err := utils.GenerateJwtToken(user.ID, params.Email, params.Role)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal membuat token!")
	}
	fmt.Println("Token: ", jwt)
	responseUser := CreateUserResponse(user)
	tx.Commit()

	return utils.JsonWithSuccess(c, responseUser, fiber.StatusCreated, "Berhasil membuat user baru!")
}
func GetAllUser (c *fiber.Ctx) error {
	users := []models.User{}
	responseUsers := []User{}

	// find all users
	database.Database.DB.Find(&users)
	for _, Users := range users {
		responseUser := CreateUserResponse(Users)
		responseUsers = append(responseUsers, responseUser)
	}
	return utils.JsonWithSuccess(c, responseUsers, fiber.StatusOK, "Berhasil mengambil semua data user!")
}
func FindIdUser (id int, user *models.User) error {
	database.Database.DB.Find(&user, "id = ?", id)
	if user.ID == 0 {
		return errors.New("the user id does not exist")
	}
	return nil
}
func UpdateUser (c *fiber.Ctx) error {
	type ParamsForUpdate struct {
		ID 		  uint 	 `json:"id"`
		Name 	  string  `json:"name"`
		Password  string `json:"password" validate:"required,min=6"`
		Email     string `json:"email" validate:"required,email"`
		Role 	  string `json:"role" validate:"required"`
	}
	var params ParamsForUpdate
	if err := c.BodyParser(&params); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memvalidasi body!")
	}

	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memvalidasi id!")
	}
	var user models.User
	if err := FindIdUser(id, &user); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memvalidasi id!")
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

	hash, err := utils.GenerateAndHashPassword(params.Password)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal melakukan hash pada password!")
	}
	// save the update here 
	user.Name = params.Name
	user.Password = hash
	user.Email = params.Email
	user.Role = params.Role

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Save(&params).Error; err != nil {
		tx.Rollback()
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseUsersUpdate := CreateUserResponse(user)
	tx.Commit()

	return utils.JsonWithSuccess(c, responseUsersUpdate, fiber.StatusOK, "Berhasil mengubah data user !")
}
func DeleteUser (c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(err.Error())
	}
	var user models.User
	if err := FindIdUser(id, &user); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memvalidasi user !")
	}
	
	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseUsers := CreateUserResponse(user)
	tx.Commit()
	
	return utils.JsonWithSuccess(c, responseUsers, fiber.StatusOK, "Berhasil menghapus data user !")
}
func Login (c *fiber.Ctx) error {
	type LoginParams struct {
		Password 	string `json:"password" validate:"required,min=6"`
		Email		string	`json:"email" validate:"required,email"`
		Role 		string  `json:"role" validate:"required"`
	}

	var params LoginParams
	if err := c.BodyParser(&params); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memvalidasi body!")
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
	// pengkondisian ketika memfilter sesuatu
	var user models.User
	if err := database.Database.DB.Where("email = ?", params.Email).First(&user).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	jwt, err := utils.GenerateJwtToken(user.ID, params.Email, params.Role)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal membuat token!")
	}
	fmt.Println("Token: ", jwt)

	//hash password yang akan dimasukkan pada saat user login
	hash, err := utils.GenerateAndHashPassword(params.Password)
	user.Password = hash
	if err := utils.CompareAndHashPassword(user.Password, params.Password); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengcompare password!")
	}
	responseLogin := CreateUserResponse(user)
	return utils.JsonWithSuccess(c, responseLogin, fiber.StatusOK, "Berhasil login!")
}
func Profile (c *fiber.Ctx) error {
	userId := c.Locals("user_id")

	var user models.User
	if err := database.Database.DB.First(&user, userId).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	responseProfile := CreateUserResponse(user)
	return utils.JsonWithSuccess(c, responseProfile, fiber.StatusOK, "Berhasil melihat profile user!")
}
func PatchUser (c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mendapatkan id!")
	}

	var user models.User
	if err := FindIdUser(id, &user); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mendapatkan id!")
	}

	inputPatchUser := make(map[string]interface{})
	if err := c.BodyParser(&inputPatchUser); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menkomvert data user!")
	}

	tx := database.Database.DB.Begin()
	defer func (){
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&user).Updates(inputPatchUser).Error; err != nil {
		tx.Rollback()
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mengupdate data!")
	}

	responseUser := CreateUserResponse(user)
	tx.Commit()

	return utils.JsonWithSuccess(c, responseUser, fiber.StatusOK, "Berhasil mengubah salah satu data!")
}
func UpdateRoleUser (c *fiber.Ctx) error {
	type ParamsUpdate struct {
		AdminPassword	string	`json:"admin_password"`
	}
	var paramsUpdate ParamsUpdate
	if err := c.BodyParser(&paramsUpdate); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memverifikasi data!")
	}

	if err := godotenv.Load(); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal meload env!")
	}

	getEnv := os.Getenv("AdminCode")
	if paramsUpdate.AdminPassword != getEnv {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Password yang dimasukkan salah!")
	}

	var user models.User
	userID := c.Locals("user_id")
	if err := database.Database.DB.First(&user, userID).Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	if err := database.Database.DB.Model(&user).Update("role", "admin").Error; err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	user.Role = "admin"

	return utils.JsonWithSuccess(c, user, fiber.StatusOK, "Berhasil mengubah role menjadi admin!")
}