package middleware

import (
	"github.com/gofiber/fiber/v2"
	"api-students/helper"
)

// RequirePermission menolak request yang role-nya tidak memiliki permission tertentu.
func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			// HANYA KIRIM PESANNYA SAJA
			return helper.Unauthorized("belum terautentikasi")
		}

		if !perms.Can(user.Role, permission) {
			// HANYA KIRIM PESANNYA SAJA
			return helper.Forbidden("role " + user.Role + " tidak memiliki hak " + permission)
		}

		return c.Next()
	}
}