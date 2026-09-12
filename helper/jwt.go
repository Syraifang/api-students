package helper

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JwtCustomClaims adalah isi "KTP" yang akan dibungkus di dalam token
type JwtCustomClaims struct {
	StudentID int    `json:"student_id"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken berfungsi mencetak token baru saat mahasiswa berhasil login
func GenerateToken(studentID int, role string) (string, error) {
	// 1. Ambil kata sandi rahasia dari file .env
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "api_students" // Cadangan kalau .env gagal terbaca
	}

	// 2. Isi data KTP-nya (aktif selama 24 jam)
	claims := &JwtCustomClaims{
		StudentID: studentID,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	// 3. Cetak dan stempel tokennya
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}