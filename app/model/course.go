package model

type Course struct {
	ID        int    `json:"id"`
	KodeMK    string `json:"kode_mk"`
	NamaMK    string `json:"nama_mk"`
	SKS       int    `json:"sks"`
	Semester  int    `json:"semester"`
	Kuota     int    `json:"kuota"`
	Terisi    int    `json:"terisi,omitempty"`
	SisaKuota int    `json:"sisa_kuota,omitempty"`
}

type CourseResponse struct {
	ID        int    `json:"id"`
	KodeMK    string `json:"kode_mk"`
	NamaMK    string `json:"nama_mk"`
	SKS       int    `json:"sks"`
	Semester  int    `json:"semester"`
	Kuota     int    `json:"kuota"`
	Terisi    int    `json:"terisi"`
	SisaKuota int    `json:"sisa_kuota"`
}

// MyCourses adalah isi field data pada GET /courses/mine
type MyCourses struct {
	Courses      []Course `json:"courses"`
	TotalCredits int      `json:"total_credits"`
	MaxCredits   int      `json:"max_credits"`
}

type SwapRequest struct {
	FromCourseID int `json:"from_course_id"`
	ToCourseID   int `json:"to_course_id"`
}

type UpdateQuotaRequest struct {
	Quota   int `json:"quota"`
	Version int `json:"version"`
}