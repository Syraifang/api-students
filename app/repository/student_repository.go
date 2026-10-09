package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data duplikat (nim atau email sudah terdaftar)")
)

// StudentRepository menyesuaikan kontrak UTS
type StudentRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	FindByUserID(ctx context.Context, userID int) (model.Student, error) // Untuk GET /auth/me
	CreateWithUser(ctx context.Context, req model.StudentCreateRequest, hashedPassword string) (model.Student, error)
	Update(ctx context.Context, id int, req model.StudentUpdateRequest) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

// 1. Helper Filter (Search, Prodi, Angkatan, dan filter Soft Delete)
func buildFilter(q model.ListQuery) (string, []any) {
	where := "WHERE deleted_at IS NULL" // Abaikan mahasiswa yang sudah di-soft delete
	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND (nim ILIKE $%d OR nama ILIKE $%d)", len(args)+1, len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}
	if q.Prodi != "" {
		where += fmt.Sprintf(" AND prodi ILIKE $%d", len(args)+1)
		args = append(args, "%"+q.Prodi+"%")
	}
	if q.Angkatan != "" {
		where += fmt.Sprintf(" AND angkatan = $%d", len(args)+1)
		args = append(args, q.Angkatan)
	}
	return where, args
}

// 2. Find All (Pagination Standar Limit Offset)
func (r *studentPostgresRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error) {
	where, args := buildFilter(q)

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("menghitung mahasiswa: %w", err)
	}

	// Aturan Sorting (nama ASC atau ipk_terakhir DESC)
	orderClause := "ORDER BY id ASC"
	switch q.Sort {
	case "nama":
		orderClause = "ORDER BY nama ASC"
	case "-ipk_terakhir":
		orderClause = "ORDER BY ipk_terakhir DESC"
	}

	sqlText := fmt.Sprintf(
		"SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at FROM students %s %s LIMIT $%d OFFSET $%d",
		where, orderClause, len(args)+1, len(args)+2,
	)
	args = append(args, q.PerPage, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar mahasiswa: %w", err)
	}
	defer rows.Close()

	hasil := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt); err != nil {
			return nil, 0, fmt.Errorf("membaca baris mahasiswa: %w", err)
		}
		hasil = append(hasil, s)
	}
	return hasil, total, rows.Err()
}

// 3. Find By ID
func (r *studentPostgresRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		"SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at FROM students WHERE id = $1 AND deleted_at IS NULL", id,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil mahasiswa: %w", err)
	}
	return s, nil
}

// 4. Create (TRANSACTION DATABASE: INSERT users lalu students)
func (r *studentPostgresRepository) CreateWithUser(ctx context.Context, req model.StudentCreateRequest, hashedPassword string) (model.Student, error) {
	// Mulai Transaksi
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Student{}, fmt.Errorf("memulai transaksi: %w", err)
	}
	// Pastikan rollback jika terjadi error sebelum Commit
	defer tx.Rollback(ctx)

	// A. Insert ke tabel users
	var userID int
	err = tx.QueryRow(ctx,
		"INSERT INTO users (email, password, role) VALUES ($1, $2, 'mahasiswa') RETURNING id",
		req.Email, hashedPassword,
	).Scan(&userID)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("insert user gagal: %w", err)
	}

	// B. Insert ke tabel students
	var s model.Student
	err = tx.QueryRow(ctx,
		"INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir",
		userID, req.NIM, req.Nama, req.Prodi, req.Angkatan, req.IPKTerakhir,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("insert student gagal: %w", err)
	}

	// C. Jika semua aman, Commit (permanen simpan ke database)
	if err := tx.Commit(ctx); err != nil {
		return model.Student{}, fmt.Errorf("commit transaksi gagal: %w", err)
	}
	return s, nil
}

// 5. Update Mahasiswa
func (r *studentPostgresRepository) Update(ctx context.Context, id int, req model.StudentUpdateRequest) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		"UPDATE students SET nama = $1, prodi = $2, angkatan = $3, ipk_terakhir = $4 WHERE id = $5 AND deleted_at IS NULL RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir",
		req.Nama, req.Prodi, req.Angkatan, req.IPKTerakhir, id,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("memperbarui mahasiswa: %w", err)
	}
	return s, nil
}

// 6. Delete (Soft Delete)
func (r *studentPostgresRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, "UPDATE students SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return fmt.Errorf("soft delete mahasiswa: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// 7. Find By User ID (Berguna untuk get profil milik sendiri di Endpoint 2)
func (r *studentPostgresRepository) FindByUserID(ctx context.Context, userID int) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		"SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at FROM students WHERE user_id = $1 AND deleted_at IS NULL", userID,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil profil mahasiswa by user_id: %w", err)
	}
	return s, nil
}

// Helper unik untuk menangkap error constraint kembar
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}