package main

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// --- Fungsi untuk memformat Respons Sukses ---
func ok(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func okList(c *fiber.Ctx, message string, data any, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(WebResponse{
		Success: true, Message: message, Data: data, Meta: meta,
	})
}

func created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location) // Memberi tahu klien di mana sumber daya baru berada
	return c.Status(fiber.StatusCreated).JSON(WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent) // 204: berhasil dihapus, tanpa body
}

// --- Fungsi untuk memformat Respons Gagal ---
func fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(WebResponse{Success: false, Message: message})
}

func failValidation(c *fiber.Ctx, errs map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(WebResponse{
		Success: false, Message: "validasi gagal", Errors: errs,
	})
}

// --- Daftar putih field yang boleh dipakai untuk mengurutkan ---
var allowedSort = map[string]bool{
	"id": true, "nim": true, "name": true, "grade": true,
}

// --- Fungsi untuk membaca Query String dari URL ---
func parseListQuery(c *fiber.Ctx) ListQuery {
	q := ListQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:   c.Query("sort", "id"),
		Order:  strings.ToLower(c.Query("order", "asc")),
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	// Batas atas wajib ada agar server tidak kehabisan memori jika diminta jutaan data
	if q.Limit > 100 {
		q.Limit = 100
	}

	// Cek daftar putih (whitelist) untuk mencegah SQL Injection nantinya
	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}

	if q.Order != "desc" {
		q.Order = "asc"
	}

	// Membaca filter boolean "is_active" jika dikirim
	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}

	return q
}