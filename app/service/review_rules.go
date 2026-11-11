package service

import "strings"

// NormalizeTags membersihkan spasi tepi, mengubah ke huruf kecil, membuang string kosong,
// dan menghapus duplikat sambil mempertahankan urutan kemunculan pertama.
func NormalizeTags(tags []string) []string {
	result := []string{}
	seen := make(map[string]bool)
	for _, raw := range tags {
		clean := strings.ToLower(strings.TrimSpace(raw))
		if clean == "" || seen[clean] {
			continue
		}
		seen[clean] = true
		result = append(result, clean)
	}
	return result
}