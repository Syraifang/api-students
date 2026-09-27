package config

import (
	"errors" // Ditambahkan untuk penanganan error
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	
	"api-students/app/model" // Ditambahkan untuk model.ErrorResponse
	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
	"api-students/route"
)

// Tambahkan perms *helper.PermissionSet di akhir parameter
func NewApp(logger *slog.Logger, pool *pgxpool.Pool, studentService *service.StudentService, perms *helper.PermissionSet) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Praktikum Backend Lanjut"),
		ErrorHandler: newErrorHandler(logger),
	})

	middleware.Register(app, logger)
	
	route.Register(app, pool, studentService, perms)

	// POIN 3: Penampung terakhir untuk URL yang tidak dikenal.
	// Menggunakan helper.NotFound sesuai instruksi modul.
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})

	return app
}

// POIN 1 & 2: newErrorHandler adalah SATU-SATUNYA tempat error berubah menjadi
// response HTTP di seluruh aplikasi. FUNGSI YANG LAMA SUDAH DIHAPUS.
func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		requestID, _ := c.Locals("requestid").(string)
		
		var appErr *helper.AppError
		switch {
		case errors.As(err, &appErr):
			// Kegagalan yang sudah kita rencanakan.
		case errors.Is(err, fiber.ErrRequestEntityTooLarge):
			appErr = &helper.AppError{
				Status:  fiber.StatusRequestEntityTooLarge,
				Code:    "PAYLOAD_TOO_LARGE",
				Message: "ukuran body melebihi batas yang diizinkan",
			}
		default:
			// Kegagalan yang tidak kita duga.
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{
					Message: fiberErr.Message,
					Status:  fiberErr.Code, 
					Code:    "HTTP_ERROR",
				}
			} else {
				appErr = helper.Internal(err)
			}
		}

		// Hanya kegagalan sisi server yang dicatat sebagai Error.
		if appErr.Status >= fiber.StatusInternalServerError {
    		logger.Error("request_failed",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status),
				slog.String("error", appErr.Unwrap().Error()))
		} else {
    		logger.Warn("request_rejected",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status))
		}

		return c.Status(appErr.Status).JSON(model.ErrorResponse{
			Success:   false,
			Code:      appErr.Code,
			Message:   appErr.Message,
			Fields:    appErr.Fields,
			RequestID: requestID,
		})
	}
}