package service

import (
	"github.com/gofiber/fiber/v2"
	"api-students/app/model"
)

// SwaggerDocs adalah kumpulan dummy handler yang memuat seluruh anotasi Swagger
// untuk endpoint Modul 12 (C.2) dan Tugas Mandiri (D.3).
// swag membaca anotasi dari blok komentar di atas fungsi ini.
type SwaggerDocs struct{}

// Dummy reference agar compiler Go tidak komplain "imported and not used"
var _ = &model.WebResponse{}

// ==========================================
// 1. AUTH ENDPOINTS
// ==========================================

// Register godoc
//
//	@Summary		Mendaftarkan akun baru
//	@Description	Role akun baru selalu "user". Password minimal 8 karakter.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.RegisterRequest	true	"Data akun baru"
//	@Success		201		{object}	model.WebResponse{data=model.User}
//	@Failure		400		{object}	model.ErrorResponse	"Body bukan JSON yang valid"
//	@Failure		409		{object}	model.ErrorResponse	"Email/username sudah dipakai"
//	@Failure		422		{object}	model.ErrorResponse	"Validasi gagal"
//	@Router			/auth/register [post]
func (d *SwaggerDocs) Register(c *fiber.Ctx) error { return nil }

// Login godoc
//
//	@Summary		Login dan mendapatkan pasangan token
//	@Description	Dibatasi 5 kali gagal per menit per IP.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.LoginRequest	true	"Email dan password"
//	@Success		200		{object}	model.WebResponse{data=model.TokenPair}
//	@Failure		401		{object}	model.ErrorResponse	"Email atau password salah"
//	@Failure		429		{object}	model.ErrorResponse	"Terlalu banyak percobaan login"
//	@Router			/auth/login [post]
func (d *SwaggerDocs) Login(c *fiber.Ctx) error { return nil }

// Refresh godoc
//
//	@Summary		Menukar refresh token dengan pasangan token baru
//	@Description	Refresh token lama langsung dicabut (rotation).
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.RefreshRequest	true	"Refresh token"
//	@Success		200		{object}	model.WebResponse{data=model.TokenPair}
//	@Failure		401		{object}	model.ErrorResponse	"Refresh token tidak valid atau kedaluwarsa"
//	@Router			/auth/refresh [post]
func (d *SwaggerDocs) Refresh(c *fiber.Ctx) error { return nil }

// Logout godoc
//
//	@Summary		Logout dan mencabut refresh token
//	@Tags			auth
//	@Security		BearerAuth
//	@Success		200	{object}	model.WebResponse
//	@Failure		401	{object}	model.ErrorResponse
//	@Router			/auth/logout [post]
func (d *SwaggerDocs) Logout(c *fiber.Ctx) error { return nil }

// Me godoc
//
//	@Summary		Mendapatkan profil pengguna yang sedang login
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.WebResponse{data=model.AuthUser}
//	@Failure		401	{object}	model.ErrorResponse
//	@Router			/auth/me [get]
func (d *SwaggerDocs) Me(c *fiber.Ctx) error { return nil }

// ==========================================
// 2. COURSES & ENROLLMENTS ENDPOINTS
// ==========================================

// ListCourses godoc
//
//	@Summary	Daftar seluruh kelas
//	@Tags		courses
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	model.WebResponse{data=[]model.Course}
//	@Failure	401	{object}	model.ErrorResponse
//	@Router		/courses [get]
func (d *SwaggerDocs) ListCourses(c *fiber.Ctx) error { return nil }

// MineCourses godoc
//
//	@Summary	KRS milik mahasiswa yang sedang login
//	@Tags		courses
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	model.WebResponse{data=model.MyCourses}
//	@Failure	401	{object}	model.ErrorResponse
//	@Router		/courses/mine [get]
func (d *SwaggerDocs) MineCourses(c *fiber.Ctx) error { return nil }

// GetCourse godoc
//
//	@Summary	Detail satu kelas
//	@Tags		courses
//	@Produce	json
//	@Param		id	path		int	true	"ID kelas"
//	@Security	BearerAuth
//	@Success	200	{object}	model.WebResponse{data=model.Course}
//	@Failure	400	{object}	model.ErrorResponse	"ID bukan angka positif"
//	@Failure	404	{object}	model.ErrorResponse
//	@Router		/courses/{id} [get]
func (d *SwaggerDocs) GetCourse(c *fiber.Ctx) error { return nil }

// EnrollCourse godoc
//
//	@Summary		Mendaftar ke sebuah kelas
//	@Description	Mengambil satu kursi dan mencatat event enrollment.created ke outbox dalam SATU transaction.
//	@Description	Kode error 409: COURSE_FULL, ALREADY_ENROLLED, CREDIT_LIMIT_EXCEEDED.
//	@Tags			courses
//	@Produce		json
//	@Param			id	path		int	true	"ID kelas"
//	@Security		BearerAuth
//	@Success		201	{object}	model.WebResponse{data=model.Course}
//	@Failure		401	{object}	model.ErrorResponse
//	@Failure		404	{object}	model.ErrorResponse
//	@Failure		409	{object}	model.ErrorResponse	"Kuota penuh, sudah terdaftar, atau batas SKS terlampaui"
//	@Router			/courses/{id}/enroll [post]
func (d *SwaggerDocs) EnrollCourse(c *fiber.Ctx) error { return nil }

// DropCourse godoc
//
//	@Summary	Keluar dari sebuah kelas
//	@Tags		courses
//	@Param		id	path	int	true	"ID kelas"
//	@Security	BearerAuth
//	@Success	204	"Berhasil keluar; tanpa body"
//	@Failure	401	{object}	model.ErrorResponse
//	@Failure	409	{object}	model.ErrorResponse	"NOT_ENROLLED"
//	@Router		/courses/{id}/enroll [delete]
func (d *SwaggerDocs) DropCourse(c *fiber.Ctx) error { return nil }

// SwapEnrollment godoc
//
//	@Summary		Pindah kelas
//	@Description	Keluar dari kelas lama dan masuk ke kelas baru sebagai satu tindakan. Bila kelas baru penuh, mahasiswa tetap di kelas lama.
//	@Tags			courses
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.SwapRequest	true	"Kelas asal dan kelas tujuan"
//	@Security		BearerAuth
//	@Success		200		{object}	model.WebResponse{data=model.Course}
//	@Failure		401		{object}	model.ErrorResponse
//	@Failure		409		{object}	model.ErrorResponse	"COURSE_FULL, NOT_ENROLLED, CREDIT_LIMIT_EXCEEDED"
//	@Failure		422		{object}	model.ErrorResponse
//	@Router			/enrollments/swap [post]
func (d *SwaggerDocs) SwapEnrollment(c *fiber.Ctx) error { return nil }

// ==========================================
// 3. REVIEWS ENDPOINTS
// ==========================================

// CreateReview godoc
//
//	@Summary		Menulis ulasan untuk sebuah kelas
//	@Description	Hanya peserta kelas. Satu mahasiswa hanya boleh menulis satu ulasan per kelas.
//	@Tags			reviews
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"ID kelas"
//	@Param			body	body		model.CreateReviewRequest	true	"Isi ulasan"
//	@Security		BearerAuth
//	@Success		201	{object}	model.WebResponse{data=model.Review}
//	@Failure		403	{object}	model.ErrorResponse	"NOT_ENROLLED"
//	@Failure		409	{object}	model.ErrorResponse	"ALREADY_REVIEWED"
//	@Failure		422	{object}	model.ErrorResponse
//	@Router			/courses/{id}/reviews [post]
func (d *SwaggerDocs) CreateReview(c *fiber.Ctx) error { return nil }

// ListReviewsByCourse godoc
//
//	@Summary		Daftar ulasan sebuah kelas, terbaru lebih dulu
//	@Description	Cursor pagination: kirim meta.next_before sebagai parameter before untuk halaman berikutnya.
//	@Tags			reviews
//	@Produce		json
//	@Param			id		path		int		true	"ID kelas"
//	@Param			tag		query		string	false	"Saring berdasarkan tag"
//	@Param			before	query		string	false	"Cursor: id ulasan terakhir yang sudah dilihat"
//	@Param			limit	query		int		false	"Jumlah per halaman, 1 sampai 50"	default(10)
//	@Security		BearerAuth
//	@Success		200	{object}	model.ReviewPage
//	@Failure		400	{object}	model.ErrorResponse
//	@Router			/courses/{id}/reviews [get]
func (d *SwaggerDocs) ListReviewsByCourse(c *fiber.Ctx) error { return nil }

// ReviewSummary godoc
//
//	@Summary	Ringkasan ulasan sebuah kelas (D.3)
//	@Tags		reviews
//	@Produce	json
//	@Param		id	path		int	true	"ID kelas"
//	@Security	BearerAuth
//	@Success	200	{object}	model.WebResponse{data=model.ReviewSummary}
//	@Failure	404	{object}	model.ErrorResponse
//	@Router		/courses/{id}/reviews/summary [get]
func (d *SwaggerDocs) ReviewSummary(c *fiber.Ctx) error { return nil }

// GetReviewByID godoc
//
//	@Summary	Detail ulasan berdasarkan ID (D.3)
//	@Tags		reviews
//	@Produce	json
//	@Param		id	path		string	true	"ID ulasan"
//	@Security	BearerAuth
//	@Success	200	{object}	model.WebResponse{data=model.Review}
//	@Failure	404	{object}	model.ErrorResponse
//	@Router		/reviews/{id} [get]
func (d *SwaggerDocs) GetReviewByID(c *fiber.Ctx) error { return nil }

// DeleteReviewByID godoc
//
//	@Summary	Menghapus ulasan (D.3)
//	@Tags		reviews
//	@Param		id	path	string	true	"ID ulasan"
//	@Security	BearerAuth
//	@Success	200	{object}	model.WebResponse
//	@Failure	404	{object}	model.ErrorResponse
//	@Router		/reviews/{id} [delete]
func (d *SwaggerDocs) DeleteReviewByID(c *fiber.Ctx) error { return nil }

// ==========================================
// 4. NOTIFICATIONS & ADMIN ENDPOINTS
// ==========================================

// MineNotifications godoc
//
//	@Summary	Notifikasi milik mahasiswa yang sedang login
//	@Tags		notifications
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	model.WebResponse{data=[]model.Notification}
//	@Failure	401	{object}	model.ErrorResponse
//	@Router		/notifications/mine [get]
func (d *SwaggerDocs) MineNotifications(c *fiber.Ctx) error { return nil }

// MonitorQueue godoc
//
//	@Summary		Keadaan outbox dan queue (admin)
//	@Description	Membutuhkan permission queue:monitor. Tetap menjawab 200 saat Redis mati; field queue_error menjelaskan sebabnya.
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	model.WebResponse{data=model.QueueStats}
//	@Failure		403	{object}	model.ErrorResponse
//	@Router			/admin/queue [get]
func (d *SwaggerDocs) MonitorQueue(c *fiber.Ctx) error { return nil }

// ==========================================
// 5. STUDENTS / USERS ENDPOINTS (D.3)
// ==========================================

// ListStudents godoc
//
//	@Summary	Daftar seluruh mahasiswa (D.3)
//	@Tags		students
//	@Produce	json
//	@Param		page	query		int		false	"Halaman"
//	@Param		per_page query		int		false	"Jumlah per halaman"
//	@Param		prodi	query		string	false	"Filter Program Studi"
//	@Param		search	query		string	false	"Pencarian Nama / NIM"
//	@Security	BearerAuth
//	@Success	200	{object}	model.WebResponse{data=[]model.Student}
//	@Failure	401	{object}	model.ErrorResponse
//	@Router		/students [get]
func (d *SwaggerDocs) ListStudents(c *fiber.Ctx) error { return nil }

// CreateStudent godoc
//
//	@Summary	Menambah data mahasiswa baru (D.3)
//	@Tags		students
//	@Accept		json
//	@Produce	json
//	@Param		body	body		model.StudentCreateRequest	true	"Data Mahasiswa"
//	@Security	BearerAuth
//	@Success	201		{object}	model.WebResponse{data=model.Student}
//	@Failure	400		{object}	model.ErrorResponse
//	@Failure	422		{object}	model.ErrorResponse
//	@Router		/students [post]
func (d *SwaggerDocs) CreateStudent(c *fiber.Ctx) error { return nil }

// GetStudent godoc
//
//	@Summary	Detail data mahasiswa (D.3)
//	@Tags		students
//	@Produce	json
//	@Param		id	path		int	true	"ID Mahasiswa"
//	@Security	BearerAuth
//	@Success	200	{object}	model.WebResponse{data=model.StudentDetailResponse}
//	@Failure	404	{object}	model.ErrorResponse
//	@Router		/students/{id} [get]
func (d *SwaggerDocs) GetStudent(c *fiber.Ctx) error { return nil }

// UpdateStudent godoc
//
//	@Summary	Mengubah data mahasiswa (D.3)
//	@Tags		students
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int							true	"ID Mahasiswa"
//	@Param		body	body		model.StudentUpdateRequest	true	"Data Mahasiswa Baru"
//	@Security	BearerAuth
//	@Success	200		{object}	model.WebResponse{data=model.Student}
//	@Failure	400		{object}	model.ErrorResponse
//	@Failure	404		{object}	model.ErrorResponse
//	@Router		/students/{id} [put]
func (d *SwaggerDocs) UpdateStudent(c *fiber.Ctx) error { return nil }

// DeleteStudent godoc
//
//	@Summary	Menghapus data mahasiswa (D.3)
//	@Tags		students
//	@Param		id	path	int	true	"ID Mahasiswa"
//	@Security	BearerAuth
//	@Success	200	{object}	model.WebResponse
//	@Failure	404	{object}	model.ErrorResponse
//	@Router		/students/{id} [delete]
func (d *SwaggerDocs) DeleteStudent(c *fiber.Ctx) error { return nil }