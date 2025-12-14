package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/ArkaniLoveCoding/fiber-project/app/router"
	"github.com/ArkaniLoveCoding/fiber-project/database"
)

func main () {
	database.ConnectionDB()
	app := fiber.New()
	
	router.HandleUserRoutes(app)
	router.HandleOrderRouter(app)
	router.HandleProductRoutes(app)
	router.HandleCheckoutRouter(app)
	router.HandlePaymentsRoutes(app)
	router.HandleStockLogsRoutes(app)

	log.Fatal(app.Listen(":9000"))
}