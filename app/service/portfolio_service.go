package service

import (
	"context"
	"errors"
	"strconv"

	"api-students/app/model"
	"api-students/app/repository"

	"github.com/gofiber/fiber/v2"
)

type PortfolioService struct {
	portfolioRepo repository.PortfolioRepository
	studentRepo   repository.StudentRepository
}

func NewPortfolioService(portfolioRepo repository.PortfolioRepository, studentRepo repository.StudentRepository) *PortfolioService {
	return &PortfolioService{
		portfolioRepo: portfolioRepo,
		studentRepo:   studentRepo,
	}
}

// checkStudentExists memvalidasi keberadaan mahasiswa di PostgreSQL
func (s *PortfolioService) checkStudentExists(ctx context.Context, studentID int64) error {
	_, err := s.studentRepo.FindByID(ctx, int(studentID))
	if err != nil {
		return errors.New("mahasiswa tidak ditemukan di PostgreSQL")
	}
	return nil
}

// GET /students/:id/portfolio
func (s *PortfolioService) GetPortfolio(c *fiber.Ctx) error {
	idParam := c.Params("id")
	studentID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID mahasiswa tidak valid",
		})
	}

	if err := s.checkStudentExists(c.UserContext(), studentID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Mahasiswa tidak ditemukan di PostgreSQL",
		})
	}

	portfolio, err := s.portfolioRepo.GetByStudentID(c.UserContext(), studentID)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Portofolio mahasiswa belum dibuat",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(portfolio)
}

// PUT /students/:id/portfolio
func (s *PortfolioService) UpsertPortfolio(c *fiber.Ctx) error {
	idParam := c.Params("id")
	studentID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID mahasiswa tidak valid",
		})
	}

	if err := s.checkStudentExists(c.UserContext(), studentID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Mahasiswa tidak ditemukan di PostgreSQL",
		})
	}

	var req model.Portfolio
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request body tidak valid",
		})
	}

	req.StudentID = studentID
	err = s.portfolioRepo.UpsertPortfolio(c.UserContext(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	res, err := s.portfolioRepo.GetByStudentID(c.UserContext(), studentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(res)
}

// POST /students/:id/portfolio/projects
func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	idParam := c.Params("id")
	studentID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID mahasiswa tidak valid",
		})
	}

	if err := s.checkStudentExists(c.UserContext(), studentID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Mahasiswa tidak ditemukan di PostgreSQL",
		})
	}

	var req model.Project
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Format request body proyek tidak valid",
		})
	}

	err = s.portfolioRepo.AddProject(c.UserContext(), studentID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Portofolio belum ada, buat portofolio terlebih dahulu",
			})
		}
		if errors.Is(err, repository.ErrProjectLimitExceeded) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Batas maksimal 20 proyek secara atomik telah tercapai",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Proyek berhasil ditambahkan",
		"project": req,
	})
}

// DELETE /students/:id/portfolio/projects/:projectId
func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	idParam := c.Params("id")
	studentID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID mahasiswa tidak valid",
		})
	}

	if err := s.checkStudentExists(c.UserContext(), studentID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Mahasiswa tidak ditemukan di PostgreSQL",
		})
	}

	projectID := c.Params("projectId")

	err = s.portfolioRepo.DeleteProject(c.UserContext(), studentID, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) || errors.Is(err, repository.ErrProjectNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Proyek atau portofolio tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Proyek berhasil dihapus",
	})
}

// GET /portfolios/stats/technologies
func (s *PortfolioService) GetTopTechnologies(c *fiber.Ctx) error {
	stats, err := s.portfolioRepo.GetTopTechnologies(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": stats,
	})
}