package service

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type EnrollmentService struct {
	enrollmentRepo repository.EnrollmentRepository
	studentRepo    repository.StudentRepository
	db             *pgxpool.Pool
}

func NewEnrollmentService(enrollmentRepo repository.EnrollmentRepository, studentRepo repository.StudentRepository, db *pgxpool.Pool) *EnrollmentService {
	return &EnrollmentService{
		enrollmentRepo: enrollmentRepo,
		studentRepo:    studentRepo,
		db:             db,
	}
}

// Create menangani endpoint POST /api/v1/enrollments (Ambil KRS)
func (s *EnrollmentService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

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

	// 2. Cari data student berdasarkan user_id
	student, err := s.studentRepo.FindByUserID(ctx, userID)
	if err != nil {
		return helper.Fail(c, fiber.StatusForbidden, "akun ini bukan mahasiswa atau data mahasiswa tidak ditemukan")
	}

	// 3. Parse request body
	var req model.EnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.BadRequest("format input tidak valid")
	}
	if errs := helper.ValidateStruct(req); errs != nil {
		return helper.Validation(errs)
	}

	// 4. Mulai Transaksi Database
	tx, err := s.enrollmentRepo.BeginTx(ctx)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memulai transaksi")
	}
	defer tx.Rollback(ctx)

	// 5. Kunci baris mata kuliah & cek kuota (Row Locking)
	course, terisi, err := s.enrollmentRepo.LockAndGetCourse(ctx, tx, req.CourseID)
	if err != nil {
		return err
	}

	if terisi >= course.Kuota {
		return helper.BadRequest("maaf, kuota mata kuliah sudah penuh")
	}

	// 6. Cek duplikasi (apakah sudah mengambil MK yang sama di tahun akademik ini)
	isDuplicate, err := s.enrollmentRepo.CheckDuplicate(ctx, student.ID, req.CourseID, req.TahunAkademik)
	if err != nil {
		return err
	}
	if isDuplicate {
		return helper.BadRequest("anda sudah mengambil mata kuliah ini pada tahun akademik tersebut")
	}

	// 7. Hitung total SKS yang sudah diambil di semester/tahun akademik tersebut
	currentTotalSKS, err := s.enrollmentRepo.GetTotalSKSCurrentSemester(ctx, student.ID, req.TahunAkademik)
	if err != nil {
		return err
	}

	// 8. Tentukan batas maksimal SKS berdasarkan IPK Terakhir
	var maxSKS int
	if student.IPKTerakhir >= 3.00 {
		maxSKS = 24
	} else if student.IPKTerakhir >= 2.50 {
		maxSKS = 21
	} else {
		maxSKS = 18
	}

	if currentTotalSKS+course.SKS > maxSKS {
		return helper.BadRequest("total SKS melebihi batas maksimal berdasarkan IPK terakhir Anda (" + strconv.Itoa(maxSKS) + " SKS)")
	}

	// 9. Simpan enrollment ke database
	err = s.enrollmentRepo.Create(ctx, tx, student.ID, req.CourseID, req.TahunAkademik)
	if err != nil {
		return err
	}

	// Commit transaksi
	if err := tx.Commit(ctx); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan KRS")
	}

	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: "mata kuliah berhasil diambil",
		Data: fiber.Map{
			"course_id":      course.ID,
			"nama_mk":        course.NamaMK,
			"sks":            course.SKS,
			"tahun_akademik": req.TahunAkademik,
			"total_sks_kini": currentTotalSKS + course.SKS,
			"batas_sks":      maxSKS,
		},
	})
}

// Delete menangani endpoint DELETE /api/v1/enrollments/:id (Batalkan KRS)
func (s *EnrollmentService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

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

	// 2. Cari data student berdasarkan user_id
	student, err := s.studentRepo.FindByUserID(ctx, userID)
	if err != nil {
		return helper.Fail(c, fiber.StatusForbidden, "akun ini bukan mahasiswa")
	}

	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return helper.BadRequest("ID enrollment tidak valid")
	}

	err = s.enrollmentRepo.Delete(ctx, id, student.ID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: "mata kuliah berhasil dibatalkan dari KRS",
	})
}