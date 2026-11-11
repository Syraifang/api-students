// [PERBAIKAN BUG 4 - compiler]: Menggunakan package service (white-box test)
// bukan package service_test karena fungsi translateReviewError tidak di-export.
package service

import (
	"errors"
	"fmt"
	"testing"

	"api-students/app/repository"
	"api-students/helper"
)

func TestTranslateReviewError(t *testing.T) {
	tests := []struct {
		err        error
		wantStatus int
		wantCode   string
	}{
		{repository.ErrInvalidID, 400, "BAD_REQUEST"},
		{repository.ErrNotFound, 404, "NOT_FOUND"},
		{repository.ErrDuplicate, 409, "ALREADY_REVIEWED"},
		{repository.ErrNotOwner, 403, "FORBIDDEN"},
		{repository.ErrOwnReview, 409, "OWN_REVIEW"},
		{repository.ErrAlreadyVoted, 409, "ALREADY_VOTED"},
		{repository.ErrReplyLimit, 409, "REPLY_LIMIT"},
		{fmt.Errorf("lapisan luar: %w", repository.ErrDuplicate), 409, "ALREADY_REVIEWED"},
		{errors.New("koneksi MongoDB putus"), 500, "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			var appErr *helper.AppError
			if !errors.As(translateReviewError(tt.err), &appErr) {
				t.Fatalf("hasil bukan *helper.AppError")
			}
			if appErr.Status != tt.wantStatus || appErr.Code != tt.wantCode {
				t.Errorf("dapat %d %s, seharusnya %d %s", appErr.Status, appErr.Code, tt.wantStatus, tt.wantCode)
			}
		})
	}
}