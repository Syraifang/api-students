package service

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

type QueueInspector interface {
	GetQueueInfo(queue string) (*asynq.QueueInfo, error)
}

type NotificationService struct {
	notifications repository.NotificationRepository
	outbox        repository.OutboxRepository
	inspector     QueueInspector
}

func NewNotificationService(
	notifications repository.NotificationRepository,
	outbox        repository.OutboxRepository,
	inspector     QueueInspector,
) *NotificationService {
	return &NotificationService{notifications: notifications, outbox: outbox, inspector: inspector}
}

// GET /notifications/mine
func (s *NotificationService) Mine(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// Mengambil user_id menyesuaikan gaya kodemu di enrollment_service
	rawUserID := c.Locals("user_id")
	if rawUserID == nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	var userID int
	switch v := rawUserID.(type) {
	case float64:
		userID = int(v)
	case int:
		userID = v
	default:
		return helper.Fail(c, fiber.StatusUnauthorized, "format sesi pengguna tidak valid")
	}

	list, err := s.notifications.FindByUser(ctx, userID, 20)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, err.Error())
	}
	return helper.Success(c, fiber.StatusOK, "notifikasi berhasil diambil", list)
}

// GET /admin/queue
func (s *NotificationService) Monitor(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	pending, published, oldest, err := s.outbox.Stats(ctx)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, err.Error())
	}

	sent, err := s.notifications.Count(ctx)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, err.Error())
	}

	stats := model.QueueStats{
		OutboxPending:     pending,
		OutboxPublished:   published,
		OldestPendingAge:  oldest.Round(time.Millisecond).String(),
		NotificationsSent: sent,
	}

	info, err := s.inspector.GetQueueInfo(model.QueueNotifications)
	if err != nil {
		stats.QueueError = "queue tidak dapat dibaca: Redis tidak dapat dihubungi atau queue belum pernah dipakai"
	} else {
		stats.Queue = &model.QueueInfo{
			Name: info.Queue, Pending: info.Pending, Active: info.Active,
			Scheduled: info.Scheduled, Retry: info.Retry, Archived: info.Archived,
			Completed: info.Completed, Processed: info.Processed, Failed: info.Failed,
		}
	}

	return helper.Success(c, fiber.StatusOK, "keadaan queue berhasil diambil", stats)
}