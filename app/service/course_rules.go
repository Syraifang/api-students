package service

const MaxCreditsPerSemester = 24

// WithinCreditLimit memeriksa apakah penambahan SKS masih dalam batas maksimal 24 SKS.
func WithinCreditLimit(taken, adding int) bool {
	return taken+adding <= MaxCreditsPerSemester
}

// CreditsAfterSwap menghitung total SKS setelah mahasiswa berpindah kelas.
func CreditsAfterSwap(taken, leaving, joining int) int {
	return taken - leaving + joining
}