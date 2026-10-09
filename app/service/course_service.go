package service

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type CourseService struct {
	courseRepo repository.CourseRepository
}

func NewCourseService(courseRepo repository.CourseRepository) *CourseService {
	return &CourseService{courseRepo: courseRepo}
}

// GetCourses menangani endpoint GET /api/v1/courses langsung dengan fiber.Ctx
func (s *CourseService) GetCourses(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// Tangkap parameter query: semester, search, available (true/false)
	var semester *int
	if semStr := c.Query("semester"); semStr != "" {
		if semVal, err := strconv.Atoi(semStr); err == nil {
			semester = &semVal
		}
	}

	var search *string
	if sStr := c.Query("search"); sStr != "" {
		search = &sStr
	}

	availableOnly := c.Query("available") == "true"

	courses, err := s.courseRepo.FindAll(ctx, semester, search, availableOnly)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: "Daftar mata kuliah berhasil diambil",
		Data:    courses,
	})
}