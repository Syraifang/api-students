package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"api-students/app/service" // Sesuaikan nama module jika beda
	"api-students/helper"      // Sesuaikan nama module jika beda
	"api-students/middleware"  // Sesuaikan nama module jika beda
	"api-students/app/repository"
)

func Register(app *fiber.App, pool *pgxpool.Pool, studentService *service.StudentService) {
	api := app.Group("/api/v1")
	
	api.Get("/health", healthCheck(pool))

	students := api.Group("/students", middleware.RequireJSON)
	students.Get("/", studentService.List)
	students.Get("/:id", studentService.Get)
	students.Post("/", studentService.Create)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)
	students.Delete("/:id", studentService.Delete)
	students.Get("/:id/prestasi", studentService.GetPrestasi)

	// 1. Rakit StudentRepo dan AuthService di sini menggunakan 'pool' yang ada
	studentRepo := repository.NewStudentRepository(pool)
	authService := service.NewAuthService(studentRepo) 

	// 2. Sambungkan rute keamanan ke grup 'api' yang sudah ada
	auth := api.Group("/auth")
	auth.Post("/register", authService.Register)
	auth.Post("/login", authService.Login)
}

// healthCheck melaporkan kondisi layanan beserta databasenya.
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