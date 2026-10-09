package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
	"api-students/helper"
)

type EnrollmentRepository interface {
	Create(ctx context.Context, tx pgx.Tx, studentID int, courseID int, tahunAkademik string) error
	Delete(ctx context.Context, id int, studentID int) error
	FindByID(ctx context.Context, id int) (*model.Enrollment, error)
	GetTotalSKSCurrentSemester(ctx context.Context, studentID int, tahunAkademik string) (int, error)
	CheckDuplicate(ctx context.Context, studentID int, courseID int, tahunAkademik string) (bool, error)
	LockAndGetCourse(ctx context.Context, tx pgx.Tx, courseID int) (*model.Course, int, error)
	BeginTx(ctx context.Context) (pgx.Tx, error)
}

type enrollmentRepository struct {
	db *pgxpool.Pool
}

func NewEnrollmentRepository(db *pgxpool.Pool) EnrollmentRepository {
	return &enrollmentRepository{db: db}
}

func (r *enrollmentRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

func (r *enrollmentRepository) CheckDuplicate(ctx context.Context, studentID int, courseID int, tahunAkademik string) (bool, error) {
	query := `SELECT COUNT(*) FROM enrollments WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3`
	var count int
	err := r.db.QueryRow(ctx, query, studentID, courseID, tahunAkademik).Scan(&count)
	return count > 0, err
}

func (r *enrollmentRepository) GetTotalSKSCurrentSemester(ctx context.Context, studentID int, tahunAkademik string) (int, error) {
	query := `
		SELECT COALESCE(SUM(c.sks), 0)
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.student_id = $1 AND e.tahun_akademik = $2
	`
	var totalSKS int
	err := r.db.QueryRow(ctx, query, studentID, tahunAkademik).Scan(&totalSKS)
	return totalSKS, err
}

// LockAndGetCourse mengunci baris mata kuliah untuk mencegah race condition kuota penuh
func (r *enrollmentRepository) LockAndGetCourse(ctx context.Context, tx pgx.Tx, courseID int) (*model.Course, int, error) {
	// Ambil data course dengan FOR UPDATE
	courseQuery := `SELECT id, kode_mk, nama_mk, sks, semester, kuota FROM courses WHERE id = $1 FOR UPDATE`
	var c model.Course
	err := tx.QueryRow(ctx, courseQuery, courseID).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, helper.NotFound("mata kuliah tidak ditemukan")
		}
		return nil, 0, err
	}

	// Hitung jumlah yang sudah terisi
	countQuery := `SELECT COUNT(*) FROM enrollments WHERE course_id = $1`
	var terisi int
	err = tx.QueryRow(ctx, countQuery, courseID).Scan(&terisi)
	if err != nil {
		return nil, 0, err
	}

	return &c, terisi, nil
}

func (r *enrollmentRepository) Create(ctx context.Context, tx pgx.Tx, studentID int, courseID int, tahunAkademik string) error {
	query := `INSERT INTO enrollments (student_id, course_id, tahun_akademik) VALUES ($1, $2, $3)`
	_, err := tx.Exec(ctx, query, studentID, courseID, tahunAkademik)
	return err
}

func (r *enrollmentRepository) Delete(ctx context.Context, id int, studentID int) error {
	query := `DELETE FROM enrollments WHERE id = $1 AND student_id = $2`
	result, err := r.db.Exec(ctx, query, id, studentID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return helper.NotFound("enrollment tidak ditemukan atau bukan milik Anda")
	}
	return nil
}

func (r *enrollmentRepository) FindByID(ctx context.Context, id int) (*model.Enrollment, error) {
	query := `SELECT id, student_id, course_id, tahun_akademik, created_at FROM enrollments WHERE id = $1`
	var e model.Enrollment
	err := r.db.QueryRow(ctx, query, id).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, helper.NotFound("enrollment tidak ditemukan")
		}
		return nil, err
	}
	return &e, nil
}