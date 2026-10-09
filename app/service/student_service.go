package service

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
	"golang.org/x/crypto/bcrypt"
)

type StudentService struct {
	repo repository.StudentRepository
}

func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	currentUser, ok := helper.CurrentUser(c)
	if !ok || currentUser.Role != "admin" {
		return helper.Forbidden("hanya admin yang berhak mengakses daftar mahasiswa")
	}

	q := helper.ParseListQuery(c)
	format, _ := helper.Negotiate(c, helper.FormatJSON, helper.FormatCSV)

	rows, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Internal(err)
	}

	if format == helper.FormatCSV {
		return helper.WriteStudentsCSV(c, rows)
	}

	lastPage := int(math.Ceil(float64(total) / float64(q.PerPage)))
	if lastPage == 0 {
		lastPage = 1
	}

	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: "Data mahasiswa berhasil diambil",
		Data:    rows,
		Meta: &model.Meta{
			CurrentPage: q.Page,
			PerPage:     q.PerPage,
			Total:       total,
			LastPage:    lastPage,
		},
	})
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id tidak valid")
	}

	currentUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Unauthorized("belum terautentikasi")
	}

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	// Aturan: Mahasiswa hanya boleh melihat datanya sendiri
	if currentUser.Role == "mahasiswa" && currentUser.UserID != student.UserID {
		return helper.Forbidden("tidak berhak mengakses data mahasiswa lain")
	}

	// TODO Endpoint 5: Ambil data MataKuliahDiambil, TotalSKS, BatasSKS dari EnrollmentRepository nanti

	return helper.Success(c, fiber.StatusOK, "data ditemukan", student)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	currentUser, ok := helper.CurrentUser(c)
	if !ok || currentUser.Role != "admin" {
		return helper.Forbidden("hanya admin yang dapat menambah mahasiswa")
	}

	var req model.StudentCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("format JSON tidak valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// UTS: Password awal mahasiswa adalah NIM yang di-hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NIM), bcrypt.DefaultCost)
	if err != nil {
		return helper.Internal(err)
	}

	newStudent, err := s.repo.CreateWithUser(ctx, req, string(hashedPassword))
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Validation(map[string]string{"nim/email": "NIM atau Email sudah terdaftar"})
		}
		return helper.Internal(err)
	}

	return helper.Created(c, "mahasiswa berhasil dibuat", newStudent, "/api/v1/students/"+strconv.Itoa(newStudent.ID))
}

func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id tidak valid")
	}

	currentUser, ok := helper.CurrentUser(c)
	if !ok || currentUser.Role != "admin" {
		return helper.Forbidden("hanya admin yang dapat mengubah mahasiswa")
	}

	var req model.StudentUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("format JSON tidak valid")
	}

	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	result, err := s.repo.Update(ctx, id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.Success(c, fiber.StatusOK, "data mahasiswa berhasil diperbarui", result)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.BadRequest("id tidak valid")
	}

	currentUser, ok := helper.CurrentUser(c)
	if !ok || currentUser.Role != "admin" {
		return helper.Forbidden("hanya admin yang dapat menghapus mahasiswa")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.NotFound("mahasiswa tidak ditemukan")
		}
		return helper.Internal(err)
	}

	return helper.NoContent(c)
}