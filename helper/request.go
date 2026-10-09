package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"api-students/app/model"
)

func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	q := model.ListQuery{
		Page:     c.QueryInt("page", 1),
		PerPage:  c.QueryInt("per_page", 10),
		Search:   strings.TrimSpace(c.Query("search")),
		Prodi:    strings.TrimSpace(c.Query("prodi")),
		Angkatan: strings.TrimSpace(c.Query("angkatan")),
		Sort:     c.Query("sort", "nama"), // Default dari UTS
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 {
		q.PerPage = 10
	}
	if q.PerPage > 50 {
		q.PerPage = 50 // Maksimal 50 sesuai soal UTS
	}
	if q.Sort != "nama" && q.Sort != "-ipk_terakhir" {
		q.Sort = "nama"
	}

	return q
}