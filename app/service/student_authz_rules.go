package service

import (
	"api-students/app/model"
	"api-students/helper"
)

// CanAccessStudent memutuskan apakah user berhak menyentuh data student tertentu.
// Sesuai prinsip C.2: Pemilik asli SELALU diizinkan, selain itu butuh permission :any.
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	// 1. Jalur Kepemilikan (Ownership)
	if current.UserID == ownerID {
		return true
	}

	// 2. Jalur Hak Akses Khusus (misalnya admin punya "student:update:any")
	return perms.Can(current.Role, anyPermission)
}