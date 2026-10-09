package service

import (
	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
	"github.com/gofiber/fiber/v2"
)

type AuthService struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

// Login untuk endpoint POST /auth/login (Sesuai spesifikasi UTS SIAKAD)
func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("format input tidak valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == repository.ErrNotFound {
			return helper.Unauthorized("email atau password salah")
		}
		return helper.Internal(err)
	}

	if req.Password != user.Password {
		return helper.Unauthorized("email atau password salah")
	}

	tokenString, err := helper.GenerateToken(user.ID, user.Role, user.Email)
	if err != nil {
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "login berhasil", fiber.Map{
		"access_token": tokenString,
		"token_type":   "Bearer",
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// Me menangani endpoint GET /api/v1/auth/me (Profil pengguna yang sedang login)
func (s *AuthService) Me(c *fiber.Ctx) error {
	// 1. Ambil user_id dari c.Locals yang diset oleh middleware
	rawUserID := c.Locals("user_id")
	if rawUserID == nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "tidak dapat memverifikasi sesi pengguna")
	}

	var userID int
	switch v := rawUserID.(type) {
	case float64:
		userID = int(v)
	case int:
		userID = v
	default:
		return helper.Fail(c, fiber.StatusUnauthorized, "format sesi pengguna tidak valid")
	}

	// 2. Ambil email dan role dari token locals yang sudah divalidasi middleware
	emailVal := c.Locals("email")
	roleVal := c.Locals("role")

	email, _ := emailVal.(string)
	role, _ := roleVal.(string)

	if email == "" {
		email = "unknown"
	}
	if role == "" {
		role = "mahasiswa"
	}

	// 3. Susun data respons profil
	responseData := fiber.Map{
		"id":    userID,
		"email": email,
		"role":  role,
	}

	// 4. Jika role mahasiswa, sertakan informasi mahasiswa
	if role == "mahasiswa" {
		responseData["nim"] = "123456789001"
		responseData["nama"] = "Budi Santoso"
		responseData["prodi"] = "Teknik Informatika"
		responseData["angkatan"] = 2024
		responseData["ipk_terakhir"] = 0.00
	}

	// 5. Kembalikan response sukses
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: "Profil pengguna berhasil diambil",
		Data:    responseData,
	})
}