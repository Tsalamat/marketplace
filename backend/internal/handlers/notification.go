package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"student-marketplace/internal/middleware"
	"student-marketplace/internal/models"
)

type NotificationHandler struct {
	db *gorm.DB
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{db: db}
}

// GET /api/v1/notifications
func (h *NotificationHandler) List(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	page  := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 30)
	offset := (page - 1) * limit

	var notifs []models.Notification
	h.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&notifs)

	var unread int64
	h.db.Model(&models.Notification{}).Where("user_id = ? AND is_read = false", userID).Count(&unread)

	return c.JSON(fiber.Map{
		"data":         notifs,
		"unread_count": unread,
	})
}

// PATCH /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkRead(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	id := c.Params("id")
	h.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true)

	return c.JSON(fiber.Map{"ok": true})
}

// POST /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllRead(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	h.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true)

	return c.JSON(fiber.Map{"ok": true})
}

// DELETE /api/v1/notifications/:id
func (h *NotificationHandler) Delete(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Notification{})
	return c.SendStatus(fiber.StatusNoContent)
}
