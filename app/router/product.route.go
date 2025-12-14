package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ArkaniLoveCoding/fiber-project/app/controller"
	"github.com/ArkaniLoveCoding/fiber-project/middleware"
)

func HandleProductRoutes(app *fiber.App) error {
	apiRoutes := app.Group("/api/v1")
	// product endpoints
	apiRoutes.Post("/product", middleware.MiddlewareRoleOnly, controller.CreatenNewProduct)
	apiRoutes.Get("/product", middleware.AuthMiddleware, controller.GetAllProducts)
	apiRoutes.Get("/product/price", middleware.AuthMiddleware, controller.GetProductPriceHighToLow)
	apiRoutes.Get("/product/stock", middleware.AuthMiddleware, controller.GetProductStockHighToLow)
	apiRoutes.Get("/product/:id", middleware.AuthMiddleware, controller.GetOneProduct)
	apiRoutes.Put("/product/:id", middleware.MiddlewareRoleOnly, controller.UpdateProduct)
	apiRoutes.Delete("/product/:id", middleware.MiddlewareRoleOnly, controller.DeleteProduct)
	apiRoutes.Get("/product/cari/search", middleware.AuthMiddleware, controller.SeacrhProductFromStatus)
	apiRoutes.Patch("/product/:id", middleware.MiddlewareRoleOnly, controller.PatchProduct)
	apiRoutes.Delete("/product/all/:id", middleware.MiddlewareRoleOnly, controller.BatchAllDeleteProduct)
	apiRoutes.Put("/product/all/:id", middleware.MiddlewareRoleOnly, controller.BatchAllUpdateProduct)

	return nil
}