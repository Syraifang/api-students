package model

import "time"

// Student adalah entitas utama yang sekarang terhubung ke PostgreSQL
type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     string    `json:"grade"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	Password  string    `json:"password,omitempty"`
	Role      string    `json:"role,omitempty"`
	OwnerID   int       `json:"owner_id"` // <-- Tambahkan baris ini
}

// CreateStudentRequest untuk metode POST (semua wajib)
type CreateStudentRequest struct {
	NIM      string `json:"nim" validate:"required,min=5,max=20"`
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Grade    string `json:"grade" validate:"required"`
	IsActive bool   `json:"is_active"`
}

// ReplaceStudentRequest untuk metode PUT (ganti seluruhnya, semua wajib)
type ReplaceStudentRequest struct {
	NIM      string `json:"nim" validate:"required,min=5,max=20"`
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Grade    string `json:"grade" validate:"required"`
	IsActive bool   `json:"is_active"`
}

// PatchStudentRequest untuk metode PATCH (ubah sebagian, pakai pointer dan omitnil)
type PatchStudentRequest struct {
	NIM      *string `json:"nim,omitempty" validate:"omitnil,min=5,max=20"`
	Name     *string `json:"name,omitempty" validate:"omitnil,min=3,max=100"`
	Grade    *string `json:"grade,omitempty" validate:"omitnil"`
	IsActive *bool   `json:"is_active,omitempty" validate:"omitnil"`
}

// Amplop baku untuk semua respons API
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// Meta untuk informasi paginasi
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ListQuery untuk menangkap parameter pencarian, pengurutan, dll
type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
}

// Offset menghitung berapa baris yang dilewati untuk halaman ini.
// Perhitungan ini pindah ke sini karena kini dipakai langsung oleh SQL.
func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}

type LoginRequest struct {
	NIM      string `json:"nim"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

// AuthUser merepresentasikan identitas pengguna yang sedang mengakses sistem.
// Dipakai oleh fungsi otorisasi Modul 6.
type AuthUser struct {
	UserID int
	Role   string
}

type ErrorResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

type Cursor struct {
	CreatedAt time.Time
	ID        int
}

type CursorQuery struct {
	Limit    int
	After    *Cursor
	Search   string
	IsActive *bool
}

type CursorMeta struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}