package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"student-marketplace/internal/middleware"
	"student-marketplace/internal/models"
)

type FriendHandler struct{ db *gorm.DB }

func NewFriendHandler(db *gorm.DB) *FriendHandler { return &FriendHandler{db: db} }

// GET /api/v1/friends — accepted friends list
func (h *FriendHandler) List(c *fiber.Ctx) error {
	me, ok := middleware.GetUserID(c)
	if !ok { return fiber.ErrUnauthorized }

	var fs []models.Friendship
	h.db.Preload("Requester.Profile").Preload("Addressee.Profile").
		Where("(requester_id = ? OR addressee_id = ?) AND status = 'accepted'", me, me).
		Find(&fs)

	// Return the "other" person for each friendship
	type FriendRow struct {
		FriendshipID string      `json:"friendship_id"`
		User         *models.User `json:"user"`
	}
	var result []FriendRow
	for _, f := range fs {
		if f.RequesterID == me {
			result = append(result, FriendRow{FriendshipID: f.ID.String(), User: f.Addressee})
		} else {
			result = append(result, FriendRow{FriendshipID: f.ID.String(), User: f.Requester})
		}
	}
	return c.JSON(result)
}

// GET /api/v1/friends/requests — pending incoming requests
func (h *FriendHandler) Requests(c *fiber.Ctx) error {
	me, ok := middleware.GetUserID(c)
	if !ok { return fiber.ErrUnauthorized }

	var fs []models.Friendship
	h.db.Preload("Requester.Profile").
		Where("addressee_id = ? AND status = 'pending'", me).
		Order("created_at DESC").Find(&fs)
	return c.JSON(fs)
}

// POST /api/v1/friends/:userId — send request
func (h *FriendHandler) Send(c *fiber.Ctx) error {
	me, ok := middleware.GetUserID(c)
	if !ok { return fiber.ErrUnauthorized }

	other := mustParseUUID(c.Params("userId"))
	if me == other { return fiber.NewError(400, "cannot add yourself") }

	var existing models.Friendship
	if h.db.Where(
		"(requester_id=? AND addressee_id=?) OR (requester_id=? AND addressee_id=?)",
		me, other, other, me).First(&existing).Error == nil {
		return fiber.NewError(409, "request already exists")
	}

	f := models.Friendship{RequesterID: me, AddresseeID: other, Status: models.FriendPending}
	h.db.Create(&f)

	h.db.Create(&models.Notification{
		UserID: other, Type: models.NotifFriendRequest,
		Title: "New friend request",
		Body:  "Someone sent you a friend request",
		Data:  map[string]any{"friendship_id": f.ID},
	})
	return c.Status(201).JSON(f)
}

// PUT /api/v1/friends/:id/accept
func (h *FriendHandler) Accept(c *fiber.Ctx) error {
	me, ok := middleware.GetUserID(c)
	if !ok { return fiber.ErrUnauthorized }

	var f models.Friendship
	if err := h.db.Where("id=? AND addressee_id=? AND status='pending'", c.Params("id"), me).
		First(&f).Error; err != nil {
		return fiber.ErrNotFound
	}
	h.db.Model(&f).Update("status", models.FriendAccepted)

	h.db.Create(&models.Notification{
		UserID: f.RequesterID, Type: models.NotifFriendRequest,
		Title: "Friend request accepted",
		Body:  "Your friend request was accepted!",
	})
	return c.JSON(fiber.Map{"status": "accepted"})
}

// PUT /api/v1/friends/:id/reject
func (h *FriendHandler) Reject(c *fiber.Ctx) error {
	me, ok := middleware.GetUserID(c)
	if !ok { return fiber.ErrUnauthorized }

	h.db.Where("id=? AND addressee_id=? AND status='pending'", c.Params("id"), me).
		Delete(&models.Friendship{})
	return c.JSON(fiber.Map{"status": "rejected"})
}

// DELETE /api/v1/friends/:userId — remove friend
func (h *FriendHandler) Remove(c *fiber.Ctx) error {
	me, ok := middleware.GetUserID(c)
	if !ok { return fiber.ErrUnauthorized }

	other := mustParseUUID(c.Params("userId"))
	h.db.Where(
		"(requester_id=? AND addressee_id=?) OR (requester_id=? AND addressee_id=?)",
		me, other, other, me).Delete(&models.Friendship{})
	return c.SendStatus(204)
}

// GET /api/v1/friends/locations — lat/lng of accepted friends
func (h *FriendHandler) Locations(c *fiber.Ctx) error {
	me, ok := middleware.GetUserID(c)
	if !ok { return fiber.ErrUnauthorized }

	var fs []models.Friendship
	h.db.Where("(requester_id=? OR addressee_id=?) AND status='accepted'", me, me).Find(&fs)

	ids := make([]interface{}, 0, len(fs))
	for _, f := range fs {
		if f.RequesterID == me {
			ids = append(ids, f.AddresseeID)
		} else {
			ids = append(ids, f.RequesterID)
		}
	}
	if len(ids) == 0 { return c.JSON([]any{}) }

	type LocRow struct {
		UserID   string  `json:"user_id"`
		Username string  `json:"username"`
		Avatar   string  `json:"avatar"`
		Lat      float64 `json:"lat"`
		Lng      float64 `json:"lng"`
	}
	var rows []LocRow
	h.db.Raw(`
		SELECT u.id::text AS user_id, u.username,
		       COALESCE(p.avatar_url,'') AS avatar,
		       COALESCE(p.lat,0) AS lat,
		       COALESCE(p.lng,0) AS lng
		FROM users u
		LEFT JOIN profiles p ON p.user_id = u.id
		WHERE u.id IN ?
		  AND p.lat IS NOT NULL AND p.lat != 0
	`, ids).Scan(&rows)
	return c.JSON(rows)
}
