package service

import (
	"testing"
	"api-students/app/model"
)

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