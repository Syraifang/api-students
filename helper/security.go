package helper

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword berfungsi mengacak password mentah menjadi teks acak yang aman untuk disimpan ke database
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash berfungsi mencocokkan password saat mahasiswa mencoba login
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}