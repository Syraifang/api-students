package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
	"api-students/route"
	"github.com/gofiber/fiber/v2"
)

// main hanya berisi urutan perakitan. Tidak ada logika bisnis,
// tidak ada query, dan tidak ada satu pun handler di sini.
func main() {
	// 1. Konfigurasi dan logger
	config.LoadEnv()
	logger := config.NewLogger()

	// 2. Database
	// -- Koneksi PostgreSQL --
	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// -- Koneksi MongoDB (TAMBAHAN BARU) --
	database.ConnectMongoDB()
	// -------------------------------------

	// 3. Perakitan dari dalam ke luar: repository -> service
	studentRepository := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository)

	// 4. Aplikasi (tanpa permissions)
	app := config.NewApp(logger, pool, studentService)
	route.Register(app, pool, studentService)

	port := config.GetEnv("APP_PORT", "3000")

	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	logger.Info("server berjalan", slog.String("port", port))

	// 5. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("sinyal berhenti diterima, menutup server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Menutup koneksi MongoDB dengan rapi (TAMBAHAN BARU)
	if err := database.MongoClient.Disconnect(ctx); err != nil {
		logger.Error("gagal menutup koneksi MongoDB", slog.String("error", err.Error()))
	}

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("gagal menutup server dengan rapi", slog.String("error", err.Error()))
	}
	
	logger.Info("server berhenti dengan rapi")

	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})
}