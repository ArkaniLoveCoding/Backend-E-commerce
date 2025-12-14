package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ArkaniLoveCoding/fiber-project/app/controller"
	"github.com/ArkaniLoveCoding/fiber-project/middleware"
)

func HandleCheckoutRouter (app *fiber.App) error {
	apiRoutes := app.Group("/api/v1")
	// order endpoints
	apiRoutes.Post("/checkout", middleware.AuthMiddleware, controller.CreateCheckout)
	apiRoutes.Get("/checkout", middleware.AuthMiddleware, controller.GetAllCheckout)
	apiRoutes.Get("/checkout/statCompanyOrder", middleware.AuthMiddleware, controller.AvarageExpenseUser)
	apiRoutes.Get("/checkout/:id", middleware.AuthMiddleware, controller.GetOneCheckout)
	apiRoutes.Put("/checkout/:id", middleware.AuthMiddleware, controller.UpdateCheckout)
	apiRoutes.Delete("/checkout/:id", middleware.AuthMiddleware, controller.DeleteCheckout)
	apiRoutes.Patch("/checkout/nominal/:id", middleware.AuthMiddleware, controller.PatchNominal)
	apiRoutes.Delete("/checkout/all/:id", middleware.AuthMiddleware, controller.BatchAllDeleteCheckout)

	return nil
}