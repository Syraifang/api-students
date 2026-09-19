package middleware

import (
	"github.com/gofiber/fiber/v2"
	"api-students/helper"
)

// RequirePermission menolak request yang role-nya tidak memiliki permission tertentu.
// Dipasang pada route yang haknya dapat diputuskan TANPA melihat isi data.
func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Mengambil data user yang disisipkan oleh middleware RequireAuth sebelumnya
		user, ok := helper.CurrentUser(c)
		if !ok {
			// Jika tidak ada identitas, berarti RequireAuth lupa dipasang. Tolak!
			return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
		}

		// Mengecek apakah role user tersebut memiliki permission yang diminta
		if !perms.Can(user.Role, permission) {
			return helper.Fail(c, fiber.StatusForbidden, "role "+user.Role+" tidak memiliki hak "+permission)
		}

		return c.Next()
	}
}