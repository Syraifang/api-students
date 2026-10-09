package helper

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"api-students/app/model"
)

const (
	FormatJSON = fiber.MIMEApplicationJSON
	FormatCSV  = "text/csv"
)

func Negotiate(c *fiber.Ctx, offered ...string) (string, error) {
	accept := strings.TrimSpace(c.Get(fiber.HeaderAccept))
	if accept == "" {
		return offered[0], nil
	}
	chosen := c.Accepts(offered...)
	if chosen == "" {
		return "", NotAcceptable("format tidak tersedia")
	}
	return chosen, nil
}

func WriteStudentsCSV(c *fiber.Ctx, students []model.Student) error {
	c.Set(fiber.HeaderContentType, FormatCSV+"; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="students.csv"`)

	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)
	if err := writer.Write([]string{"id", "nim", "nama", "prodi", "angkatan", "ipk_terakhir"}); err != nil {
		return Internal(err)
	}

	for _, u := range students {
		row := []string{
			strconv.Itoa(u.ID), u.NIM, u.Nama, u.Prodi,
			strconv.Itoa(u.Angkatan), fmt.Sprintf("%.2f", u.IPKTerakhir),
		}
		if err := writer.Write(row); err != nil {
			return Internal(err)
		}
	}
	writer.Flush()
	return c.SendString(buffer.String())
}