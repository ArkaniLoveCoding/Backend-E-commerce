package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ArkaniLoveCoding/fiber-project/models"
)



type DBInstance struct {
	DB *gorm.DB
}

var Database DBInstance

func ConnectionDB() error {
	if err := godotenv.Load(); err != nil {
		return errors.New(err.Error())
	}
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, pass, name, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("failed to initialize database, got error", err)
		return nil
	}

	log.Println("Connect into database has been succesfully!")
	db.Logger = logger.Default.LogMode(logger.Info)
	log.Println("Running into migrations")

	db.AutoMigrate(&models.User{}, &models.Product{}, &models.Order{}, &models.Checkout{}, &models.StockLog{}, &models.Payment{})

	Database = DBInstance{DB: db}
	return nil
}


func Pagination (c *fiber.Ctx)  func (db *gorm.DB) * gorm.DB  {
	return func(db *gorm.DB) *gorm.DB {
	
	paramsPage := c.Query("page", "1")
	page, err := strconv.Atoi(paramsPage)
	if err != nil {
		return nil
	}

	paramsPageSize := c.Query("limit", "10")
	limit, err := strconv.Atoi(paramsPageSize)
	if err != nil {
		return nil
	}
	if page <= 0 {
		page = 1
	}

	switch {
	case limit > 100:
		limit = 100
	case limit <= 0:
		limit = 10
	}

	offset := (page - 1) * limit
	return db.Offset(offset).Limit(limit)
	}
}