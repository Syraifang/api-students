package service

import (
	"github.com/gofiber/fiber/v2"
	"api-students/helper"
	"api-students/app/model"
	"api-students/app/repository"
)

type AuthService struct {
	repo repository.StudentRepository
}

func NewAuthService(repo repository.StudentRepository) AuthService {
	return AuthService{repo: repo}
}

// Register untuk mendaftarkan mahasiswa baru beserta passwordnya
func (s AuthService) Register(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.Student
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "format input tidak valid")
	}

	if req.NIM == "" || req.Password == "" {
		return helper.Fail(c, fiber.StatusBadRequest, "nim dan password wajib diisi")
	}

	// Acak password menggunakan helper Bcrypt sebelum disimpan
	hashedPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses password")
	}
	req.Password = hashedPassword

	// Set default role jika kosong
	if req.Role == "" {
		req.Role = "student"
	}

	// Simpan ke database
	result, err := s.repo.Create(ctx, req)
	if err != nil {
		if err == repository.ErrDuplicate {
			return helper.Fail(c, fiber.StatusConflict, "nim sudah terdaftar")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mendaftar mahasiswa")
	}

	// Kosongkan field password agar tidak bocor di response JSON Postman
	result.Password = ""

	return helper.Success(c, fiber.StatusCreated, "registrasi berhasil", result)
}

// Login untuk mengecek kecocokan data dan memberikan token JWT
func (s AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "format input tidak valid")
	}

	// 1. Cari mahasiswa berdasarkan NIM
	student, err := s.repo.FindByNIM(ctx, req.NIM)
	if err != nil {
		// Ingat: Jangan beritahu secara spesifik apakah NIM atau Password yang salah demi keamanan
		return helper.Fail(c, fiber.StatusUnauthorized, "nim atau password salah") 
	}

	// 2. Cek apakah password yang diketik cocok dengan password acak di database
	if !helper.CheckPasswordHash(req.Password, student.Password) {
		return helper.Fail(c, fiber.StatusUnauthorized, "nim atau password salah")
	}

	// 3. Buat Token JWT karena login sukses
	tokenString, err := helper.GenerateToken(student.ID, student.Role)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat sesi login")
	}

	// 4. Kirimkan token ke mahasiswa
	resp := model.AuthResponse{
		Token: tokenString,
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", resp)
}