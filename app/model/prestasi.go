package model

type Prestasi struct {
	ID           int    `json:"id"`
	StudentID    int    `json:"student_id"`
	NamaPrestasi string `json:"nama_prestasi"`
	Juara        string `json:"juara"`
}