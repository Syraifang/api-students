package service

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type ReviewCourseGate interface {
	FindByID(ctx context.Context, id int) (*model.Course, error)
	IsEnrolled(ctx context.Context, courseID, userID int) (bool, error)
}

type ReviewService struct {
	reviews repository.ReviewRepository
	courses ReviewCourseGate
}

func NewReviewService(reviews repository.ReviewRepository, courses ReviewCourseGate) *ReviewService {
	return &ReviewService{reviews: reviews, courses: courses}
}

func translateReviewError(err error) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return &helper.AppError{Status: http.StatusBadRequest, Code: "BAD_REQUEST", Message: err.Error()}
	case errors.Is(err, repository.ErrNotFound):
		return &helper.AppError{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: err.Error()}
	case errors.Is(err, repository.ErrDuplicate):
		return &helper.AppError{Status: http.StatusConflict, Code: "ALREADY_REVIEWED", Message: "Anda sudah mengulas kelas ini"}
	case errors.Is(err, repository.ErrNotOwner):
		return &helper.AppError{Status: http.StatusForbidden, Code: "FORBIDDEN", Message: err.Error()}
	case errors.Is(err, repository.ErrOwnReview):
		return &helper.AppError{Status: http.StatusConflict, Code: "OWN_REVIEW", Message: err.Error()}
	case errors.Is(err, repository.ErrAlreadyVoted):
		return &helper.AppError{Status: http.StatusConflict, Code: "ALREADY_VOTED", Message: err.Error()}
	case errors.Is(err, repository.ErrReplyLimit):
		return &helper.AppError{Status: http.StatusConflict, Code: "REPLY_LIMIT", Message: err.Error()}
	default:
		return &helper.AppError{Status: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: "terjadi kesalahan pada server"}
	}
}

func (s *ReviewService) ListByCourse(c *fiber.Ctx) error {
	courseID, err := c.ParamsInt("id")
	if err != nil || courseID <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"success": false, "code": "BAD_REQUEST", "message": "ID kelas tidak valid"})
	}
	list, err := s.reviews.FindByCourse(c.UserContext(), courseID)
	if err != nil {
		return translateReviewError(err)
	}
	return c.Status(http.StatusOK).JSON(model.ReviewPage{
		Success: true,
		Message: "daftar ulasan berhasil diambil",
		Data:    list,
		Meta:    model.ReviewPageMeta{Limit: 10, HasMore: false, NextBefore: ""},
	})
}

func (s *ReviewService) Create(c *fiber.Ctx) error {
	courseID, err := c.ParamsInt("id")
	if err != nil || courseID <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"success": false, "code": "BAD_REQUEST", "message": "ID kelas tidak valid"})
	}

	userID, _ := c.Locals("user_id").(int)

	var req model.CreateReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"success": false, "code": "BAD_REQUEST", "message": "body bukan JSON yang valid"})
	}

	if req.Rating < 1 || req.Rating > 5 || len(strings.TrimSpace(req.Comment)) < 10 {
		return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{"success": false, "code": "VALIDATION_ERROR", "message": "validasi gagal"})
	}

	if _, err := s.courses.FindByID(c.UserContext(), courseID); err != nil {
		appErr := translateReviewError(err).(*helper.AppError)
		return c.Status(appErr.Status).JSON(fiber.Map{"success": false, "code": appErr.Code, "message": appErr.Message})
	}

	enrolled, err := s.courses.IsEnrolled(c.UserContext(), courseID, userID)
	if err != nil {
		appErr := translateReviewError(err).(*helper.AppError)
		return c.Status(appErr.Status).JSON(fiber.Map{"success": false, "code": appErr.Code, "message": appErr.Message})
	}
	if !enrolled {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{"success": false, "code": "NOT_ENROLLED", "message": "bukan peserta kelas"})
	}

	created, err := s.reviews.Create(c.UserContext(), model.Review{
		CourseID: courseID,
		UserID:   userID,
		Rating:   req.Rating,
		Comment:  strings.TrimSpace(req.Comment),
		Tags:     NormalizeTags(req.Tags),
	})
	if err != nil {
		appErr := translateReviewError(err).(*helper.AppError)
		return c.Status(appErr.Status).JSON(fiber.Map{"success": false, "code": appErr.Code, "message": appErr.Message})
	}

	return c.Status(http.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: "ulasan berhasil dibuat",
		Data:    created,
	})
}