package service

import (
	"testing"
	"api-students/app/model"
)

// 1. Test Validasi POST (Create)
func TestValidateCreate(t *testing.T) {
	// Skenario gagal: Data kosong
	req := model.CreateStudentRequest{
		NIM:  "",
		Name: "",
	}
	
	errs := ValidateCreate(req)
	if len(errs) == 0 {
		t.Errorf("Diharapkan ada error validasi untuk data kosong, tapi lolos")
	}
}

// 2. Test Validasi PUT (Replace)
func TestValidateReplace(t *testing.T) {
	// Skenario gagal: Data wajib ada yang kosong
	req := model.ReplaceStudentRequest{
		NIM:   "", // Sengaja dikosongkan agar memicu error
		Name:  "",
		Grade: "A",
	}
	
	errs := ValidateReplace(req)
	if len(errs) == 0 {
		t.Errorf("Diharapkan ada error karena NIM/Nama kosong, tapi malah lolos")
	}
}

// 3. Test Penerapan PATCH (Update Sebagian)
func TestApplyPatch(t *testing.T) {
	current := model.Student{
		ID: 1, NIM: "12345", Name: "Andi", Grade: "B",
	}
	
	namaBaru := "Andi Susanto"
	req := model.PatchStudentRequest{
		Name: &namaBaru,
	}
	
	updated, _ := ApplyPatch(current, req)
	
	if updated.Name != "Andi Susanto" {
		t.Errorf("Diharapkan nama berubah menjadi Andi Susanto, tapi mendapat %s", updated.Name)
	}
	if updated.NIM != "12345" {
		t.Errorf("Diharapkan NIM tidak berubah, tapi ikut berubah")
	}
}