package repository

import (
	"context"
	"errors"

	"api-students/app/model"
)

var (
	ErrInvalidID    = errors.New("id tidak valid")
	ErrNotOwner     = errors.New("bukan pemilik data")
	ErrOwnReview    = errors.New("tidak dapat menilai ulasan sendiri")
	ErrAlreadyVoted = errors.New("sudah pernah memberi suara")
	ErrReplyLimit   = errors.New("batas balasan tercapai")
	ErrCourseFull   = errors.New("kuota kelas penuh")
)

type ReviewRepository interface {
	Create(ctx context.Context, r model.Review) (model.Review, error)
	FindByCourse(ctx context.Context, courseID int) ([]model.Review, error)
}