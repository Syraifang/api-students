package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"api-students/app/repository"
	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
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

	// Rute Courses (Semua role terautentikasi)
	courseRepo := repository.NewCourseRepository(pool)
	courseService := service.NewCourseService(courseRepo)
	api.Get("/courses", middleware.RequireJSON, middleware.RequireAuth, courseService.GetCourses)

	// Rute Students (Berpelindung Token)
	students := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth)
	
	students.Get("/", studentService.List)
	students.Post("/", studentService.Create)
	students.Get("/:id", studentService.Get)
	students.Put("/:id", studentService.Replace)
	students.Delete("/:id", studentService.Delete)

	// Rute Enrollments / KRS (Mahasiswa yang terautentikasi)
	studentRepo := repository.NewStudentRepository(pool)
	enrollmentRepo := repository.NewEnrollmentRepository(pool)
	enrollmentService := service.NewEnrollmentService(enrollmentRepo, studentRepo, pool)
	
	enrollments := api.Group("/enrollments", middleware.RequireJSON, middleware.RequireAuth)
	enrollments.Post("/", enrollmentService.Create)
	enrollments.Delete("/:id", enrollmentService.Delete)
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