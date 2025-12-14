package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ArkaniLoveCoding/fiber-project/app/controller"
	"github.com/ArkaniLoveCoding/fiber-project/middleware"
)

func HandleUserRoutes(app *fiber.App) error {
	apiRoutes := app.Group("/api/v1")

	// user endpoints
	apiRoutes.Post("/user/registration", controller.CreateUserNew)
	apiRoutes.Get("/user", middleware.AuthMiddleware, controller.GetAllUser)
	apiRoutes.Put("/user/role", middleware.MiddlewareAuthChangeRole, controller.UpdateRoleUser)
	apiRoutes.Put("/user/:id", middleware.AuthMiddleware, controller.UpdateUser)
	apiRoutes.Delete("/user/:id", middleware.AuthMiddleware, controller.DeleteUser)
	apiRoutes.Post("/user/login", middleware.AuthMiddleware, controller.Login)
	apiRoutes.Get("/user/profile", middleware.AuthMiddlewareForProfile, controller.Profile)
	apiRoutes.Patch("/user/:id", middleware.AuthMiddleware, controller.PatchUser)

	return nil
}