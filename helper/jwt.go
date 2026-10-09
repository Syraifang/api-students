package helper

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"api-students/app/model"
	"github.com/gofiber/fiber/v2"
)

// JwtCustomClaims adalah isi "KTP" yang akan dibungkus di dalam token
type JwtCustomClaims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken berfungsi mencetak token baru saat berhasil login
func GenerateToken(userID int, role string, email string) (string, error) {
	// 1. Ambil kata sandi rahasia dari file .env
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "api_students" // Cadangan kalau .env gagal terbaca
	}

	// 2. Isi data KTP-nya (aktif selama 24 jam)
	claims := &JwtCustomClaims{
		UserID: userID,
		Role:   role,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	// 3. Cetak dan stempel tokennya
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	role, okRole := c.Locals("role").(string)
	email, _ := c.Locals("email").(string)
	
	var userID int
	// Ubah pencarian dari student_id menjadi user_id
	if idFloat, ok := c.Locals("user_id").(float64); ok {
		userID = int(idFloat)
	} else if idInt, ok := c.Locals("user_id").(int); ok {
		userID = idInt
	} else {
		return model.AuthUser{}, false
	}

	if !okRole {
		return model.AuthUser{}, false
	}

	return model.AuthUser{
		UserID: userID,
		Role:   role,
		Email:  email,
	}, true
}