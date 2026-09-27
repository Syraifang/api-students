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

// Register menyimpan password TANPA hash agar bisa dilihat langsung di database
func (s AuthService) Register(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.Student
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("format input tidak valid")
	}

	if req.NIM == "" || req.Password == "" {
		return helper.BadRequest("nim dan password wajib diisi")
	}

	// TIDAK PERLU HASH, simpan langsung apa adanya supaya bisa dilihat di DB
	// req.Password = req.Password 

	if req.Role == "" {
		req.Role = "user"
	}

	result, err := s.repo.Create(ctx, req)
	if err != nil {
		if err == repository.ErrDuplicate {
			return helper.Conflict("nim sudah terdaftar")
		}
		return helper.Internal(err)
	}

	result.Password = ""
	return helper.Success(c, fiber.StatusCreated, "registrasi berhasil", result)
}

// Login mencocokkan password secara langsung (teks biasa)
func (s AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("format input tidak valid")
	}

	student, err := s.repo.FindByNIM(ctx, req.NIM)
	if err != nil {
		return helper.Unauthorized("nim atau password salah") 
	}

	// COCOKAN LANGSUNG SEBAGAI STRING BIASA
	if req.Password != student.Password {
		return helper.Unauthorized("nim atau password salah")
	}

	tokenString, err := helper.GenerateToken(student.ID, student.Role)
	if err != nil {
		return helper.Internal(err)
	}

	resp := model.AuthResponse{
		Token: tokenString,
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", resp)
}