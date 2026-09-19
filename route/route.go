package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
	"api-students/app/repository"
)

// TAMBAHKAN perms *helper.PermissionSet di sini
func Register(app *fiber.App, pool *pgxpool.Pool, studentService *service.StudentService, perms *helper.PermissionSet) {
	api := app.Group("/api/v1")
	
	api.Get("/health", healthCheck(pool))

	students := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth)
	
	// ==========================================
	// 1. Hak dapat diputuskan TANPA melihat data -> Dijaga Middleware
	// ==========================================
	students.Get("/", middleware.RequirePermission(perms, "student:list"), studentService.List)
	students.Post("/", middleware.RequirePermission(perms, "student:create"), studentService.Create)
	students.Delete("/:id", middleware.RequirePermission(perms, "student:delete"), studentService.Delete)

	// ==========================================
	// 2. Hak bergantung pada KEPEMILIKAN data -> Diperiksa di Service nanti
	// ==========================================
	students.Get("/:id", studentService.Get)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)
	students.Get("/:id/prestasi", studentService.GetPrestasi)

	studentRepo := repository.NewStudentRepository(pool)
	authService := service.NewAuthService(studentRepo) 

	auth := api.Group("/auth")
	auth.Post("/register", authService.Register)
	auth.Post("/login", authService.Login)
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