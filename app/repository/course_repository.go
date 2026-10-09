package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	
	"api-students/app/model"
	"api-students/helper"
)

type CourseRepository interface {
	FindAll(ctx context.Context, semester *int, search *string, availableOnly bool) ([]model.CourseResponse, error)
	FindByID(ctx context.Context, id int) (*model.Course, error)
}

type courseRepository struct {
	db *pgxpool.Pool
}

func NewCourseRepository(db *pgxpool.Pool) CourseRepository {
	return &courseRepository{db: db}
}

func (r *courseRepository) FindAll(ctx context.Context, semester *int, search *string, availableOnly bool) ([]model.CourseResponse, error) {
	query := `
		SELECT 
			c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
			COALESCE(en.terisi, 0) as terisi,
			(c.kuota - COALESCE(en.terisi, 0)) as sisa_kuota
		FROM courses c
		LEFT JOIN (
			SELECT course_id, COUNT(*) as terisi
			FROM enrollments
			GROUP BY course_id
		) en ON c.id = en.course_id
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if semester != nil {
		query += fmt.Sprintf(" AND c.semester = $%d", argIdx)
		args = append(args, *semester)
		argIdx++
	}

	if search != nil && *search != "" {
		query += fmt.Sprintf(" AND (c.kode_mk ILIKE $%d OR c.nama_mk ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+*search+"%")
		argIdx++
	}

	if availableOnly {
		query += ` AND (c.kuota - COALESCE(en.terisi, 0)) > 0`
	}

	query += ` ORDER BY c.kode_mk ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []model.CourseResponse
	for rows.Next() {
		var c model.CourseResponse
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi, &c.SisaKuota); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}

	return courses, nil
}

func (r *courseRepository) FindByID(ctx context.Context, id int) (*model.Course, error) {
	query := `SELECT id, kode_mk, nama_mk, sks, semester, kuota FROM courses WHERE id = $1`
	var c model.Course
	err := r.db.QueryRow(ctx, query, id).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, helper.NotFound("mata kuliah tidak ditemukan")
		}
		return nil, err
	}
	return &c, nil
}