package service

import (
	"strings"
	"api-students/app/model" // Sesuaikan jika nama module-mu berbeda
)

// ValidateCreate memeriksa isi permintaan pembuatan student baru.
func ValidateCreate(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi"
	}
	if strings.TrimSpace(req.Grade) == "" {
		errs["grade"] = "wajib diisi"
	}
	// Kamu bisa tambahkan validasi lain, misal panjang NIM atau format Grade
	return errs
}

// ValidateReplace memeriksa isi permintaan PUT (semua field wajib ada).
func ValidateReplace(req model.ReplaceStudentRequest) map[string]string {
	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi pada PUT"
	}
	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi pada PUT"
	}
	if strings.TrimSpace(req.Grade) == "" {
		errs["grade"] = "wajib diisi pada PUT"
	}
	return errs
}

// ApplyPatch menyalin field yang dikirim ke data yang sudah ada (untuk PATCH).
func ApplyPatch(current model.Student, req model.PatchStudentRequest) (model.Student, map[string]string) {
	errs := map[string]string{}
	
	if req.NIM != nil {
		if strings.TrimSpace(*req.NIM) == "" {
			errs["nim"] = "tidak boleh kosong"
		} else {
			current.NIM = *req.NIM
		}
	}
	
	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			current.Name = *req.Name
		}
	}

	if req.Grade != nil {
		if strings.TrimSpace(*req.Grade) == "" {
			errs["grade"] = "tidak boleh kosong"
		} else {
			current.Grade = *req.Grade
		}
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	
	return current, errs
}

// IsEmptyPatch menandai permintaan PATCH yang tidak mengubah apa pun.
func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

// CountTotalPages membulatkan ke atas tanpa memakai bilangan pecahan.
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
}