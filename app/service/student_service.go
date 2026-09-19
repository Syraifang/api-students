package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"api-students/app/model"      
	"api-students/app/repository" 
	"api-students/helper"         
)

// StudentService memegang peran controller dan memanggil use case
type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{
		repo:  repo,
		perms: perms,
	}
}

func (s *StudentService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	q := helper.ParseListQuery(c)

	students, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
	}

	return helper.SuccessList(c, "daftar mahasiswa berhasil diambil", students, &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: CountTotalPages(total, q.Limit),
	})
}

func (s *StudentService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// 1. Ambil ID dari URL
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// 2. Ambil identitas user yang sedang login
	currentUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	// 3. Ambil data dari database DULU untuk mengetahui OwnerID-nya
	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "gagal mengambil data mahasiswa")
	}

	// 4. CEK HAK AKSES KEPEMILIKAN
	if !CanAccessStudent(currentUser, student.OwnerID, s.perms, "student:read:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengakses data milik orang lain")
	}

	// 5. Lanjut kembalikan data sukses
	return helper.Success(c, fiber.StatusOK, "data ditemukan", student)
}

func (s *StudentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// 1. Ambil identitas user yang sedang login
	currentUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)
	req.Grade = strings.TrimSpace(req.Grade)

	if errs := ValidateCreate(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	// 2. Isi data mahasiswa baru
	newStudent, err := s.repo.Create(ctx, model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
		OwnerID:  currentUser.UserID, // Mencegah spoofing: nilai owner_id diisi otomatis dari token login
	})

	if err != nil {
		return translateError(c, err, "gagal menyimpan mahasiswa")
	}

	return helper.Created(c, "mahasiswa berhasil dibuat", newStudent, "/api/v1/students/"+strconv.Itoa(newStudent.ID))
}

func translateError(c *fiber.Ctx, err error, generalMessage string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(c, fiber.StatusConflict, "NIM sudah dipakai")
	default:
		return helper.Fail(c, fiber.StatusInternalServerError, generalMessage)
	}
}

func (s *StudentService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// 1. Ambil identitas login
	currentUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	// 2. Ambil data asli dari database
	existingStudent, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "gagal mengambil data mahasiswa")
	}

	// 3. CEK HAK AKSES KEPEMILIKAN
	if !CanAccessStudent(currentUser, existingStudent.OwnerID, s.perms, "student:update:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengubah data milik orang lain")
	}

	var req model.ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if errs := ValidateReplace(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := s.repo.Update(ctx, model.Student{
		ID:       id,
		NIM:      strings.TrimSpace(req.NIM),
		Name:     strings.TrimSpace(req.Name),
		Grade:    strings.TrimSpace(req.Grade),
		IsActive: req.IsActive,
		OwnerID:  existingStudent.OwnerID, // Pastikan OwnerID lama tidak hilang saat ditimpa
	})

	if err != nil {
		return translateError(c, err, "gagal memperbarui mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "mahasiswa berhasil diganti seluruhnya", result)
}

func (s *StudentService) Patch(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	// 1. Ambil identitas login
	currentUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if IsEmptyPatch(req) {
		return helper.Fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	// 2. Ambil data asli dari database
	existingStudent, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "gagal mengambil data mahasiswa")
	}

	// 3. CEK HAK AKSES KEPEMILIKAN
	if !CanAccessStudent(currentUser, existingStudent.OwnerID, s.perms, "student:update:any") {
		return helper.Fail(c, fiber.StatusForbidden, "tidak berhak mengubah data milik orang lain")
	}

	updated, errs := ApplyPatch(existingStudent, req)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return translateError(c, err, "gagal memperbarui mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "mahasiswa berhasil diperbarui sebagian", result)
}

func (s *StudentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(c, err, "gagal menghapus mahasiswa")
	}

	return helper.NoContent(c)
}

func (s StudentService) GetPrestasi(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return helper.Fail(c, fiber.StatusBadRequest, "id mahasiswa harus berupa angka positif")
	}

	listPrestasi, err := s.repo.FindPrestasiByStudentID(ctx, id)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data prestasi mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "data prestasi berhasil diambil", listPrestasi)
}