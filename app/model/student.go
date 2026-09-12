package model

import "time"

// Student adalah entitas utama yang sekarang terhubung ke PostgreSQL
type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     string    `json:"grade"` // Diubah menjadi string menyesuaikan VARCHAR(2)
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"` // Kolom baru dari database
	Password string `json:"password,omitempty"`
	Role     string `json:"role,omitempty"`
}

// CreateStudentRequest untuk metode POST (semua wajib)
type CreateStudentRequest struct {
	NIM      string `json:"nim"`
	Name     string `json:"name"`
	Grade    string `json:"grade"`
	IsActive bool   `json:"is_active"`
}

// ReplaceStudentRequest untuk metode PUT (ganti seluruhnya, semua wajib)
type ReplaceStudentRequest struct {
	NIM      string `json:"nim"`
	Name     string `json:"name"`
	Grade    string `json:"grade"`
	IsActive bool   `json:"is_active"`
}

// PatchStudentRequest untuk metode PATCH (ubah sebagian, pakai pointer)
type PatchStudentRequest struct {
	NIM      *string `json:"nim,omitempty"`
	Name     *string `json:"name,omitempty"`
	Grade    *string `json:"grade,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
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