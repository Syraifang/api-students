package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	
	"api-students/helper" // Pastikan ini mengarah ke folder helper milikmu
)

// RequireAuth berfungsi sebagai satpam yang mengecek tiket masuk (token JWT)
func RequireAuth(c *fiber.Ctx) error {
	// 1. Tagih tiket dari header "Authorization"
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return helper.Fail(c, fiber.StatusUnauthorized, "akses ditolak: token tidak ditemukan")
	}

	// 2. Pastikan format tiketnya adalah "Bearer <token_acak>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return helper.Fail(c, fiber.StatusUnauthorized, "akses ditolak: format token tidak valid")
	}
	tokenString := parts[1]

	// 3. Siapkan stempel rahasia (harus sama persis dengan yang di jwt.go)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "api_students"
	}

	// 4. Periksa keaslian tiket
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Pastikan metode enkripsinya benar
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return helper.Fail(c, fiber.StatusUnauthorized, "akses ditolak: token tidak valid atau sudah kadaluarsa")
	}

	// 5. Ekstrak isi KTP dari tiket, simpan ke saku aplikasi (Locals) 
	// Ini berguna kalau nanti kita butuh tahu "Siapa sih yang lagi login?"
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		c.Locals("student_id", claims["student_id"])
		c.Locals("role", claims["role"])
	}

	// 6. Jika semua aman, persilakan lewat!
	return c.Next()
}