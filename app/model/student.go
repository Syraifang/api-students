package model

import "time"

// User merepresentasikan tabel "users"
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"` // Sembunyikan password saat diubah ke JSON
	Role     string `json:"role"`
}

// Student adalah entitas utama mahasiswa sesuai tabel UTS
type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id,omitempty"` // Relasi ke tabel users
	NIM         string     `json:"nim"`               // 12 digit, uniq
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"` // Untuk fitur soft delete
}

// StudentCreateRequest untuk endpoint POST /students (menambah mahasiswa + akun)
type StudentCreateRequest struct {
	NIM         string  `json:"nim" validate:"required,len=12"`
	Nama        string  `json:"nama" validate:"required"`
	Email       string  `json:"email" validate:"required,email"`
	Prodi       string  `json:"prodi" validate:"required"`
	Angkatan    int     `json:"angkatan" validate:"required,min=2000"`
	IPKTerakhir float64 `json:"ipk_terakhir" validate:"omitempty,min=0,max=4"`
}

// StudentUpdateRequest untuk endpoint PUT /students/:id (NIM tidak ada karena dilarang ubah)
type StudentUpdateRequest struct {
	Nama        string  `json:"nama" validate:"required"`
	Prodi       string  `json:"prodi" validate:"required"`
	Angkatan    int     `json:"angkatan" validate:"required,min=2000"`
	IPKTerakhir float64 `json:"ipk_terakhir" validate:"omitempty,min=0,max=4"`
}

// LoginRequest untuk endpoint POST /auth/login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// AuthUser merepresentasikan profil pengguna di dalam JWT
type AuthUser struct {
	UserID int
	Role   string
	Email  string
}

// WebResponse menyesuaikan standar format JSON UTS
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"` // Dipakai untuk error validasi bersarang
}

// Meta menyesuaikan format pagination dari UTS
type Meta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}

// ListQuery untuk menangkap filter & pagination dari parameter URL GET /students
type ListQuery struct {
	Page     int
	PerPage  int
	Prodi    string
	Angkatan string
	Search   string // Mencari NIM atau nama
	Sort     string // "nama" atau "-ipk_terakhir"
}

// Offset menghitung limit & offset database
func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.PerPage
}

// StudentDetailResponse untuk endpoint GET /students/:id (Detail + daftar MK)
type StudentDetailResponse struct {
	Student
	MataKuliahDiambil []Course `json:"mata_kuliah_diambil"`
	TotalSKS          int      `json:"total_sks"`
	BatasSKS          int      `json:"batas_sks"`
}

type ErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}