package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/ArkaniLoveCoding/fiber-project/utils"
)

func MiddlewareRoleOnly (c *fiber.Ctx) error {
	tokenHeader := c.Get("Authorization")
	if tokenHeader == "" {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal menemukan authorization!")
	}

	tokenString := strings.TrimPrefix(tokenHeader, "Bearer ")
	if err := utils.VerifyJwt(tokenString); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	claims, err := utils.ExtractClaimsFromJWT(tokenString)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, err.Error())
	}

	resultRole := claims["role"].(string)

	if resultRole == "member" {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Hanya admin yang bisa mengubah atau mengatur ini!")
	}
	
	return c.Next()
}
func MiddlewareAuthChangeRole (c *fiber.Ctx) error {
	tokenString := c.Get("Authorization")
	if tokenString == "" {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Tidak ada token yang diverifikasi!")
	}

	token := strings.TrimPrefix(tokenString, "Bearer ")
	if err  := utils.VerifyJwt(token); err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal memverifikasi token!")
	}

	claims, err := utils.ExtractClaimsFromJWT(token)
	if err != nil {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "Gagal mendapatkan token!")
	}

	rawId, ok := claims["user_id"]
	if !ok {
		return utils.JsonWithError(c, fiber.StatusBadRequest, "User ID tidak ditemukan di token!")
	}
	userId := uint(rawId.(float64))
	c.Locals("user_id", userId)

	return c.Next()
}