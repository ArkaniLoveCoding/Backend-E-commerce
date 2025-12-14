package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ArkaniLoveCoding/fiber-project/app/controller"
	"github.com/ArkaniLoveCoding/fiber-project/middleware"
)

func HandleStockLogsRoutes(app *fiber.App) error {
	apiRoutes := app.Group("/api/v1")
	
	// stockLogo endpoints 
	apiRoutes.Post("/stock", middleware.MiddlewareRoleOnly, controller.CreateNewNote)
	return nil
}