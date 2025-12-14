package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ArkaniLoveCoding/fiber-project/utils"
)

func AuthMiddleware (c *fiber.Ctx) error {
	tokenHeader := c.Get("Authorization")
	if tokenHeader == "" {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Tidak ditemukan data yang diinginkan!")
	}
	tokenString := strings.TrimPrefix(tokenHeader, "Bearer ")
	if err := utils.VerifyJwt(tokenString); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}
	return c.Next()
}
func AuthMiddlewareForProfile(c *fiber.Ctx) error {
	tokenHeader := c.Get("Authorization")
	if tokenHeader == "" {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Token tidak ditemukan!")
	}

	tokenString := strings.TrimPrefix(tokenHeader, "Bearer ")
	if err := utils.VerifyJwt(tokenString); err != nil {
		return utils.JsonWithError(c, fiber.StatusUnauthorized, "Token tidak valid atau expired!")
	}

	claims, err := utils.ExtractClaimsFromJWT(tokenString)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal membaca token!")
	}
	rawId, ok := claims["user_id"]
	if !ok {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "User ID tidak ditemukan di token!")
	}
	userId := uint(rawId.(float64))
	c.Locals("user_id", userId)

	return c.Next()
}
