package utils

import "github.com/gofiber/fiber/v2"

func JsonWithSuccess (c *fiber.Ctx, data interface{}, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": "true",
		"data": data,
		"message": message,
	})	
}
func JsonWithPaginationSucces (c *fiber.Ctx, data interface{}, statusCode int, message string, page int, limit int) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": "true",
		"data": data,
		"message": message,
		"pagination": fiber.Map{
			"page": page,
			"limit": limit,
		},
	})
}
func JsonWithError (c *fiber.Ctx, statusCode int, message string) error {
	return  c.Status(statusCode).JSON(fiber.Map{
		"success": "false",
		"data": "Nothing!",
		"message": message,
	})
}