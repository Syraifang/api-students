package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	
	"api-students/helper"
)

// RequireAuth berfungsi sebagai satpam yang mengecek tiket masuk (token JWT)
func RequireAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return helper.Unauthorized("akses ditolak: token tidak ditemukan")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return helper.Unauthorized("akses ditolak: format token tidak valid")
	}
	tokenString := parts[1]

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "api_students"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return helper.Unauthorized("akses ditolak: token tidak valid atau sudah kadaluarsa")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		c.Locals("student_id", claims["student_id"])
		c.Locals("role", claims["role"])
	}

	return c.Next()
}