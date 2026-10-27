package model

type Project struct {
	ID        string   `json:"id"`
	Judul     string   `json:"judul"`
	Teknologi []string `json:"teknologi"`
	Tautan    string   `json:"tautan"`
	Tahun     int      `json:"tahun"`
}

type Certificate struct {
	Nama     string `json:"nama"`
	Penerbit string `json:"penerbit"`
	Tanggal  string `json:"tanggal"`
}

type Portfolio struct {
	ID           string        `json:"id,omitempty"`
	StudentID    int64         `json:"student_id"`
	Headline     string        `json:"headline"`
	Bio          string        `json:"bio"`
	Projects     []Project     `json:"projects"`
	Certificates []Certificate `json:"certificates"`
}

type TechStat struct {
	Teknologi string `json:"teknologi"`
	Jumlah    int    `json:"jumlah"`
}