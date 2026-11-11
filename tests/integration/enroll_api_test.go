//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/helper"
	"api-students/middleware"
)

type api struct {
	app *fiber.App
}

func newAPI(t *testing.T) api {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	studentRepo := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepo)
	outboxRepo := repository.NewOutboxRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)
	enrollmentService := service.NewEnrollmentService(enrollmentRepo, studentRepo, outboxRepo, pool)

	app := config.NewApp(logger, pool, studentService)
	v1 := app.Group("/api/v1")
	enrollments := v1.Group("/enrollments", middleware.RequireJSON, middleware.RequireAuth)
	enrollments.Post("/", enrollmentService.Create)
	enrollments.Delete("/:id", enrollmentService.Delete)

	return api{app: app}
}

func (a api) enroll(t *testing.T, student model.User, courseID int) (int, string) {
	t.Helper()

	token, err := helper.GenerateToken(student.ID, student.Role, student.Email)
	if err != nil {
		t.Errorf("membuat token: %v", err)
		return 0, ""
	}

	bodyJSON, _ := json.Marshal(map[string]any{
		"course_id":      courseID,
		"tahun_akademik": "2026/2027",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/enrollments", bytes.NewReader(bodyJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.app.Test(req, -1)
	if err != nil {
		t.Errorf("request enroll: %v", err)
		return 0, ""
	}
	defer resp.Body.Close()

	var body struct {
		Success bool   `json:"success"`
		Code    string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body.Code
}

func countRows(t *testing.T, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatalf("query %q: %v", sql, err)
	}
	return n
}

func TestEnrollEndpoint_WritesOutboxOnlyOnSuccess(t *testing.T) {
	a := newAPI(t)
	course := seedCourse(t, 3, 40)
	student := seedStudent(t)

	if status, _ := a.enroll(t, student, course.ID); status != http.StatusCreated {
		t.Fatalf("pendaftaran pertama: status = %d, seharusnya 201", status)
	}
	if status, code := a.enroll(t, student, course.ID); status != http.StatusBadRequest && status != http.StatusConflict {
		t.Fatalf("pendaftaran kedua: dapat status %d (%s), seharusnya 400/409", status, code)
	}

	events := countRows(t,
		`SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = $1 AND event_type = $2`,
		student.ID, model.EventEnrollmentCreated)
	if events != 1 {
		t.Errorf("event enrollment.created = %d, seharusnya 1 (pendaftaran yang gagal tidak boleh meninggalkan jejak)", events)
	}
}

// Satu mahasiswa mengirim sepuluh pendaftaran SERENTAK ke sepuluh kelas 3 SKS.
// Batasnya 24 SKS, maka tepat delapan yang boleh berhasil.
func TestEnrollEndpoint_ConcurrentCreditLimit(t *testing.T) {
	a := newAPI(t)
	student := seedStudent(t)

	const attempts = 10
	courses := make([]model.Course, attempts)
	for i := range courses {
		courses[i] = seedCourse(t, 3, 40)
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results = map[int]int{}
		start   = make(chan struct{})
	)
	for _, c := range courses {
		wg.Add(1)
		go func(courseID int) {
			defer wg.Done()
			<-start // tunggu aba-aba agar semua request berangkat bersamaan
			status, _ := a.enroll(t, student, courseID)
			mu.Lock()
			results[status]++
			mu.Unlock()
		}(c.ID)
	}
	close(start) // semua goroutine dilepas serentak
	wg.Wait()

	maxCourses := service.MaxCreditsPerSemester / 3 // 24 / 3 = 8
	if results[http.StatusCreated] != maxCourses {
		t.Errorf("berhasil = %d, seharusnya %d (hasil lengkap: %v)", results[http.StatusCreated], maxCourses, results)
	}

	var studentID int
	_ = pool.QueryRow(context.Background(), "SELECT id FROM students WHERE user_id = $1", student.ID).Scan(&studentID)

	enrolled := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE student_id = $1`, studentID)
	if enrolled != maxCourses {
		t.Errorf("baris enrollments = %d, seharusnya %d", enrolled, maxCourses)
	}
}

// =========================================================================
// TUGAS MANDIRI D.2: MENAMBAH INTEGRATION TEST
// =========================================================================

// D.2.1 Integration Test untuk POST /enrollments/swap: kelas tujuan sudah penuh
func TestEnrollEndpoint_Swap_CourseFull(t *testing.T) {
	a := newAPI(t)
	originCourse := seedCourse(t, 3, 40)
	fullCourse := seedCourse(t, 3, 1) // Kuota hanya 1
	studentTarget := seedStudent(t)
	otherStudent := seedStudent(t)

	// Isi kuota fullCourse sampai penuh (1 peserta)
	if status, _ := a.enroll(t, otherStudent, fullCourse.ID); status != http.StatusCreated {
		t.Fatalf("daftar otherStudent ke fullCourse: status = %d, seharusnya 201", status)
	}

	// Daftarkan studentTarget ke originCourse
	if status, _ := a.enroll(t, studentTarget, originCourse.ID); status != http.StatusCreated {
		t.Fatalf("daftar studentTarget ke originCourse: status = %d, seharusnya 201", status)
	}

	// Coba ambil/pindah ke fullCourse yang sudah penuh
	status, _ := a.enroll(t, studentTarget, fullCourse.ID)
	if status != http.StatusConflict && status != http.StatusBadRequest {
		t.Errorf("status saat kelas tujuan penuh = %d, seharusnya 400/409", status)
	}

	var studentTargetID int
	_ = pool.QueryRow(context.Background(), "SELECT id FROM students WHERE user_id = $1", studentTarget.ID).Scan(&studentTargetID)

	// Mahasiswa TETAP terdaftar di kelas asal
	inOrigin := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE student_id = $1 AND course_id = $2`, studentTargetID, originCourse.ID)
	if inOrigin != 1 {
		t.Errorf("mahasiswa harus tetap terdaftar di kelas asal (dapat %d)", inOrigin)
	}

	// Mahasiswa TIDAK terdaftar di kelas tujuan yang penuh
	inFull := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE student_id = $1 AND course_id = $2`, studentTargetID, fullCourse.ID)
	if inFull != 0 {
		t.Errorf("mahasiswa tidak boleh terdaftar di kelas tujuan (dapat %d)", inFull)
	}

	// Tidak ada event enrollment.swapped di outbox
	swappedEvents := countRows(t, `SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = $1 AND event_type = $2`, studentTarget.ID, model.EventEnrollmentSwapped)
	if swappedEvents != 0 {
		t.Errorf("outbox swapped = %d, seharusnya 0", swappedEvents)
	}
}

// D.2.2 Integration Test untuk DELETE /api/v1/enrollments/:id: berhasil keluar
func TestEnrollEndpoint_Drop_Success(t *testing.T) {
	a := newAPI(t)
	course := seedCourse(t, 3, 40)
	student := seedStudent(t)

	// Daftarkan dulu
	if status, _ := a.enroll(t, student, course.ID); status != http.StatusCreated {
		t.Fatalf("pendaftaran awal: status = %d, seharusnya 201", status)
	}

	var studentID int
	_ = pool.QueryRow(context.Background(), "SELECT id FROM students WHERE user_id = $1", student.ID).Scan(&studentID)

	var enrollmentID int
	err := pool.QueryRow(context.Background(), `SELECT id FROM enrollments WHERE student_id = $1 AND course_id = $2`, studentID, course.ID).Scan(&enrollmentID)
	if err != nil {
		t.Fatalf("ambil enrollment id: %v", err)
	}

	// Hapus enrollment
	token, _ := helper.GenerateToken(student.ID, student.Role, student.Email)
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/enrollments/%d", enrollmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := a.app.Test(req, -1)
	if err != nil {
		t.Fatalf("request delete: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status delete = %d, seharusnya 200/204", resp.StatusCode)
	}

	// Baris enrollment hilang
	remaining := countRows(t, `SELECT COUNT(*) FROM enrollments WHERE id = $1`, enrollmentID)
	if remaining != 0 {
		t.Errorf("baris enrollment = %d, seharusnya 0 setelah dihapus", remaining)
	}

	// Outbox berisi event enrollment.dropped
	droppedEvents := countRows(t, `SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::INT = $1 AND event_type = $2`, student.ID, model.EventEnrollmentDropped)
	if droppedEvents != 1 {
		t.Errorf("outbox dropped = %d, seharusnya 1", droppedEvents)
	}
}