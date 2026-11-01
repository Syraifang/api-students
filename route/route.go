package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/database"
	"api-students/helper"
	"api-students/middleware"
	"api-students/config"
	"github.com/hibiken/asynq"
)

// Disederhanakan tanpa parameter perms
func Register(app *fiber.App, pool *pgxpool.Pool, studentService *service.StudentService) {
	api := app.Group("/api/v1")
	api.Get("/health", healthCheck(pool))

	// Rute Auth (Publik)
	userRepo := repository.NewUserRepository(pool)
	authService := service.NewAuthService(userRepo)
	
	auth := api.Group("/auth")
	auth.Post("/login", authService.Login)
	auth.Get("/me", middleware.RequireAuth, authService.Me)

	outboxRepo := repository.NewOutboxRepository(pool)
	txManager := repository.NewTxManager(pool)

	// Rute Courses (Semua role terautentikasi)
	courseRepo := repository.NewCourseRepository(pool)
	courseService := service.NewCourseService(courseRepo, outboxRepo, txManager)
	api.Get("/courses", middleware.RequireJSON, middleware.RequireAuth, courseService.GetCourses)

	// Rute Students (Berpelindung Token)
	studentRepo := repository.NewStudentRepository(pool)
	students := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth)
	
	students.Get("/", studentService.List)
	students.Post("/", studentService.Create)

	// Rute Portofolio Mahasiswa (MongoDB - Modul 10) 
	// Dipindahkan ke atas sebelum /:id agar rute spesifik terbaca duluan
	portfolioRepo := repository.NewPortfolioRepository(database.MongoDB)
	portfolioService := service.NewPortfolioService(portfolioRepo, studentRepo)

	students.Get("/:id/portfolio", portfolioService.GetPortfolio)
	students.Put("/:id/portfolio", portfolioService.UpsertPortfolio)
	students.Post("/:id/portfolio/projects", portfolioService.AddProject)
	students.Delete("/:id/portfolio/projects/:projectId", portfolioService.DeleteProject)

	students.Get("/:id", studentService.Get)
	students.Put("/:id", studentService.Replace)
	students.Delete("/:id", studentService.Delete)

	api.Get("/portfolios/stats/technologies", middleware.RequireJSON, middleware.RequireAuth, portfolioService.GetTopTechnologies)

	// Rute Enrollments / KRS (Mahasiswa yang terautentikasi)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)
	enrollmentService := service.NewEnrollmentService(enrollmentRepo, studentRepo, outboxRepo, pool)
	
	enrollments := api.Group("/enrollments", middleware.RequireJSON, middleware.RequireAuth)
	enrollments.Post("/", enrollmentService.Create)
	enrollments.Delete("/:id", enrollmentService.Delete)

	notificationRepo := repository.NewNotificationRepository(pool)
	inspector := asynq.NewInspector(config.RedisOpt())
	notificationService := service.NewNotificationService(notificationRepo, outboxRepo, inspector)

	notifications := api.Group("/notifications", middleware.RequireJSON, middleware.RequireAuth)
	notifications.Get("/mine", notificationService.Mine)

	admin := api.Group("/admin", middleware.RequireJSON, middleware.RequireAuth)
	admin.Get("/queue", notificationService.Monitor)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}