package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ArkaniLoveCoding/fiber-project/app/controller"
	"github.com/ArkaniLoveCoding/fiber-project/middleware"
)

func HandleOrderRouter (app *fiber.App) error {
	apiRoutes := app.Group("/api/v1")
	// order endpoints
	apiRoutes.Post("/order", middleware.AuthMiddleware, controller.CreateOrder)
	apiRoutes.Get("/order", middleware.AuthMiddleware, controller.GetAllOrder)
	apiRoutes.Get("/order/product/search", middleware.MiddlewareRoleOnly, controller.FindOrderWithId)
	apiRoutes.Get("/order/findProductFromUser", middleware.MiddlewareRoleOnly, controller.FindOrderUserToProduct)
	apiRoutes.Get("/order/findUser", middleware.MiddlewareRoleOnly, controller.FindUserOrder)
	apiRoutes.Get("/order/sumQty", middleware.MiddlewareRoleOnly, controller.FindOrderUserToProduct)
	apiRoutes.Get("/order/sumAndOrder", middleware.MiddlewareRoleOnly, controller.CountAndSumOrderAndQuantity)
	apiRoutes.Get("/order/totalOrderForManyUser", middleware.MiddlewareRoleOnly, controller.TotalOrderForManyUsers)
	apiRoutes.Get("/order/totalQtyProduct", middleware.MiddlewareRoleOnly, controller.TotalQtyProduct)
	apiRoutes.Get("/order/productStat", middleware.MiddlewareRoleOnly, controller.ProductsStatistikToOrder)
	apiRoutes.Get("/order/:id", middleware.AuthMiddleware, controller.GetOneOrder)
	apiRoutes.Put("/order/:id", middleware.AuthMiddleware, controller.UpdateOrder)
	apiRoutes.Delete("/order/:id", middleware.AuthMiddleware, controller.DeleteOrder)
	apiRoutes.Get("/order/cari/search", middleware.AuthMiddleware, controller.SearchProductAndUser)
	apiRoutes.Patch("/order/quantity/:id", middleware.AuthMiddleware, controller.PatchQuantityOrder)
	apiRoutes.Delete("/order/all/:id", middleware.AuthMiddleware, controller.BatchAllDeleteOrder)
	apiRoutes.Put("/order/all/:id", middleware.AuthMiddleware, controller.BatchAllUpdateOrder)

	return nil
}