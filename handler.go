package main

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Tempat penyimpanan data sementara (karena belum pakai database sungguhan)
var students []Student
var nextID = 1

// --- Fungsi Bantuan Internal ---
func findStudentIndex(id int) int {
	for i := range students {
		if students[i].ID == id {
			return i
		}
	}
	return -1
}

func matchSearch(s Student, keyword string) bool {
	keyword = strings.ToLower(keyword)
	return strings.Contains(strings.ToLower(s.Name), keyword) ||
		strings.Contains(strings.ToLower(s.NIM), keyword)
}

func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

func isNIMExists(nim string, ignoreID int) bool {
	for _, s := range students {
		if s.ID != ignoreID && strings.EqualFold(s.NIM, nim) {
			return true
		}
	}
	return false
}

// --- 1. GET /students (Daftar & Paginasi) ---
func listStudents(c *fiber.Ctx) error {
	q := parseListQuery(c)

	// Saring (Filter & Search)
	hasil := []Student{}
	for _, s := range students {
		if q.IsActive != nil && s.IsActive != *q.IsActive {
			continue
		}
		if q.Search != "" && !matchSearch(s, q.Search) {
			continue
		}
		hasil = append(hasil, s)
	}

	// Urutkan (Sort)
	sort.SliceStable(hasil, func(i, j int) bool {
		var lebihKecil bool
		switch q.Sort {
		case "nim":
			lebihKecil = hasil[i].NIM < hasil[j].NIM
		case "name":
			lebihKecil = hasil[i].Name < hasil[j].Name
		case "grade":
			lebihKecil = hasil[i].Grade < hasil[j].Grade
		default:
			lebihKecil = hasil[i].ID < hasil[j].ID
		}
		if q.Order == "desc" {
			return !lebihKecil
		}
		return lebihKecil
	})

	// Potong sesuai halaman (Paginasi)
	total := len(hasil)
	totalPages := (total + q.Limit - 1) / q.Limit
	mulai := (q.Page - 1) * q.Limit
	if mulai > total {
		mulai = total
	}
	akhir := mulai + q.Limit
	if akhir > total {
		akhir = total
	}

	return okList(c, "daftar mahasiswa berhasil diambil", hasil[mulai:akhir], &Meta{
		Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages,
	})
}

// --- 2. GET /students/:id (Ambil Satu) ---
func getStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	return ok(c, "mahasiswa ditemukan", students[i])
}

// --- 3. POST /students (Tambah Baru) ---
func createStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Name = strings.TrimSpace(req.Name)

	if req.NIM == "" {
		errs["nim"] = "wajib diisi"
	} else if isNIMExists(req.NIM, 0) {
		return fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
	}

	if req.Name == "" {
		errs["name"] = "wajib diisi"
	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "nilai harus antara 0 dan 100"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	baru := Student{
		ID:       nextID,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}
	students = append(students, baru)
	nextID++

	return created(c, "mahasiswa berhasil dibuat", baru, "/api/v1/students/"+strconv.Itoa(baru.ID))
}

// --- 4. PUT /students/:id (Ganti Seluruhnya) ---
func replaceStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	var req ReplaceStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}
	if strings.TrimSpace(req.NIM) == "" {
		errs["nim"] = "wajib diisi pada PUT"
	} else if isNIMExists(req.NIM, id) {
		return fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
	}

	if strings.TrimSpace(req.Name) == "" {
		errs["name"] = "wajib diisi pada PUT"
	}
	if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "nilai harus antara 0 dan 100 pada PUT"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	students[i].NIM = req.NIM
	students[i].Name = req.Name
	students[i].Grade = req.Grade
	students[i].IsActive = req.IsActive

	return ok(c, "data mahasiswa berhasil diganti seluruhnya", students[i])
}

// --- 5. PATCH /students/:id (Ubah Sebagian) ---
func patchStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	errs := map[string]string{}

	if req.NIM != nil {
		val := strings.TrimSpace(*req.NIM)
		if val == "" {
			errs["nim"] = "tidak boleh kosong"
		} else if isNIMExists(val, id) {
			return fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
		} else {
			students[i].NIM = val
		}
	}

	if req.Name != nil {
		val := strings.TrimSpace(*req.Name)
		if val == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			students[i].Name = val
		}
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errs["grade"] = "nilai harus antara 0 dan 100"
		} else {
			students[i].Grade = *req.Grade
		}
	}

	if req.IsActive != nil {
		students[i].IsActive = *req.IsActive
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	return ok(c, "data mahasiswa berhasil diperbarui sebagian", students[i])
}

// --- 6. DELETE /students/:id ---
func deleteStudent(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	i := findStudentIndex(id)
	if i == -1 {
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	students = append(students[:i], students[i+1:]...)
	return noContent(c)
}