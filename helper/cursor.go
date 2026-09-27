package helper

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"api-students/app/model"
	"github.com/gofiber/fiber/v2"
)

var (
	ErrInvalidCursor = errors.New("cursor tidak sah")
)

func EncodeCursor(createdAt time.Time, id int) string {
	raw := strconv.FormatInt(createdAt.UTC().UnixNano(), 10) + "|" + strconv.Itoa(id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(encoded string) (model.Cursor, error) {
	if strings.TrimSpace(encoded) == "" {
		return model.Cursor{}, ErrInvalidCursor
	}
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return model.Cursor{}, ErrInvalidCursor
	}
	parts := strings.Split(string(decoded), "|")
	if len(parts) != 2 {
		return model.Cursor{}, ErrInvalidCursor
	}
	nanos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return model.Cursor{}, ErrInvalidCursor
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil || id < 1 {
		return model.Cursor{}, ErrInvalidCursor
	}
	return model.Cursor{CreatedAt: time.Unix(0, nanos).UTC(), ID: id}, nil
}

func ParseCursorQuery(c *fiber.Ctx) (model.CursorQuery, error) {
	limit := c.QueryInt("limit", 10)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	search := strings.TrimSpace(c.Query("search"))
	
	var isActive *bool
	if activeStr := c.Query("is_active"); activeStr != "" {
		val := activeStr == "true"
		isActive = &val
	}

	var after *model.Cursor
	if cursorStr := c.Query("cursor"); cursorStr != "" {
		decoded, err := DecodeCursor(cursorStr)
		if err != nil {
			return model.CursorQuery{}, BadRequest("cursor tidak valid")
		}
		after = &decoded
	}

	return model.CursorQuery{
		Limit:    limit,
		After:    after,
		Search:   search,
		IsActive: isActive,
	}, nil
}

func SuccessCursor(c *fiber.Ctx, message string, data any, meta any) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    meta,
	})
}