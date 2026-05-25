package handlers

import (
	"student-marketplace/internal/middleware"
	"student-marketplace/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ChatHandler struct {
	db *gorm.DB
}

func NewChatHandler(db *gorm.DB) *ChatHandler {
	return &ChatHandler{db: db}
}

// List all chats for current user
func (h *ChatHandler) ListChats(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var participations []models.ChatParticipant
	h.db.Where("user_id = ?", userID).Find(&participations)

	chatIDs := make([]interface{}, len(participations))
	for i, p := range participations {
		chatIDs[i] = p.ChatID
	}

	if len(chatIDs) == 0 {
		return c.JSON([]models.Chat{})
	}

	var chats []models.Chat
	h.db.
		Preload("Participants.User.Profile").
		Where("id IN ?", chatIDs).
		Order("updated_at DESC").
		Find(&chats)

	// Attach last message and unread count
	for i := range chats {
		var lastMsg models.ChatMessage
		if err := h.db.Preload("Sender.Profile").
			Where("chat_id = ? AND is_deleted = false", chats[i].ID).
			Order("created_at DESC").
			First(&lastMsg).Error; err == nil {
			chats[i].LastMessage = &lastMsg
		}

		var unread int64
		// Find last_read for current user
		var part models.ChatParticipant
		h.db.Where("chat_id = ? AND user_id = ?", chats[i].ID, userID).First(&part)
		if part.LastRead != nil {
			h.db.Model(&models.ChatMessage{}).
				Where("chat_id = ? AND sender_id != ? AND created_at > ? AND is_deleted = false",
					chats[i].ID, userID, part.LastRead).
				Count(&unread)
		} else {
			h.db.Model(&models.ChatMessage{}).
				Where("chat_id = ? AND sender_id != ? AND is_deleted = false", chats[i].ID, userID).
				Count(&unread)
		}
		chats[i].UnreadCount = int(unread)
	}

	return c.JSON(chats)
}

// Get chat messages
func (h *ChatHandler) GetMessages(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	chatID := c.Params("chatId")

	// Verify participant
	var part models.ChatParticipant
	if err := h.db.Where("chat_id = ? AND user_id = ?", chatID, userID).First(&part).Error; err != nil {
		return fiber.ErrForbidden
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 50)
	offset := (page - 1) * limit

	var messages []models.ChatMessage
	h.db.Preload("Sender.Profile").
		Where("chat_id = ? AND is_deleted = false", chatID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&messages)

	// Reverse for chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return c.JSON(messages)
}

// Start / get direct chat with a user
func (h *ChatHandler) GetOrCreateDirect(c *fiber.Ctx) error {
	myID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	otherUserIDStr := c.Params("userId")
	otherUserID := mustParseUUID(otherUserIDStr)

	// Validate IDs
	if myID == otherUserID {
		return fiber.NewError(fiber.StatusBadRequest, "cannot start chat with yourself")
	}

	// Verify other user exists
	var otherUser models.User
	if err := h.db.First(&otherUser, "id = ?", otherUserID).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	// Find existing direct chat between these two users
	var existingChatID struct{ ChatID string }
	h.db.Raw(`
		SELECT cp1.chat_id
		FROM chat_participants cp1
		JOIN chat_participants cp2 ON cp1.chat_id = cp2.chat_id
		JOIN chats ch ON ch.id = cp1.chat_id
		WHERE cp1.user_id = ?
		  AND cp2.user_id = ?
		  AND ch.is_group = false
		  AND ch.order_id IS NULL
		LIMIT 1
	`, myID, otherUserID).Scan(&existingChatID)

	if existingChatID.ChatID != "" {
		var chat models.Chat
		h.db.Preload("Participants.User.Profile").First(&chat, "id = ?", existingChatID.ChatID)
		return c.JSON(chat)
	}

	// Create new chat
	chat := models.Chat{IsGroup: false}
	h.db.Create(&chat)
	h.db.Create(&models.ChatParticipant{ChatID: chat.ID, UserID: myID})
	h.db.Create(&models.ChatParticipant{ChatID: chat.ID, UserID: otherUserID})

	h.db.Preload("Participants.User.Profile").First(&chat, chat.ID)
	return c.Status(fiber.StatusCreated).JSON(chat)
}

// POST /api/v1/chat/group - Create a group chat
func (h *ChatHandler) CreateGroupChat(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var req struct {
		Name      string   `json:"name" validate:"required,min=1,max=100"`
		MemberIDs []string `json:"member_ids" validate:"required,min=1,max=100"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Validate member IDs and check users exist
	memberUUIDs := make(map[string]bool)
	memberUUIDs[userID.String()] = true // Add creator

	for _, idStr := range req.MemberIDs {
		if idStr == userID.String() {
			continue // Skip duplicate
		}
		memberID := mustParseUUID(idStr)
		var user models.User
		if err := h.db.First(&user, "id = ?", memberID).Error; err != nil {
			return fiber.NewError(fiber.StatusNotFound, "user "+idStr+" not found")
		}
		memberUUIDs[idStr] = true
	}

	// Create group chat
	chat := models.Chat{
		IsGroup: true,
		Name:    req.Name,
	}
	if err := h.db.Create(&chat).Error; err != nil {
		return fiber.ErrInternalServerError
	}

	// Add all members
	for memberIDStr := range memberUUIDs {
		memberID := mustParseUUID(memberIDStr)
		h.db.Create(&models.ChatParticipant{ChatID: chat.ID, UserID: memberID})
	}

	h.db.Preload("Participants.User.Profile").First(&chat, chat.ID)
	return c.Status(fiber.StatusCreated).JSON(chat)
}

// POST /api/v1/chat/:chatId/members/:userId - Add member to group chat
func (h *ChatHandler) AddMember(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	chatID := c.Params("chatId")
	newMemberIDStr := c.Params("userId")
	newMemberID := mustParseUUID(newMemberIDStr)

	// Verify current user is member
	var userPart models.ChatParticipant
	if err := h.db.Where("chat_id = ? AND user_id = ?", chatID, userID).First(&userPart).Error; err != nil {
		return fiber.ErrForbidden
	}

	// Verify chat is group chat
	var chat models.Chat
	if err := h.db.First(&chat, "id = ? AND is_group = true", chatID).Error; err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "not a group chat")
	}

	// Check if new member exists
	var newMember models.User
	if err := h.db.First(&newMember, "id = ?", newMemberID).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	// Check if already a member
	var existingPart models.ChatParticipant
	if h.db.Where("chat_id = ? AND user_id = ?", chatID, newMemberID).First(&existingPart).Error == nil {
		return fiber.NewError(fiber.StatusBadRequest, "user already in chat")
	}

	// Add member
	h.db.Create(&models.ChatParticipant{ChatID: chat.ID, UserID: newMemberID})

	h.db.Preload("Participants.User.Profile").First(&chat, chat.ID)
	return c.JSON(chat)
}

// DELETE /api/v1/chat/:chatId/members/:userId - Remove member from group chat
func (h *ChatHandler) RemoveMember(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	chatID := c.Params("chatId")
	removeMemberIDStr := c.Params("userId")
	removeMemberID := mustParseUUID(removeMemberIDStr)

	// Verify current user is member
	var userPart models.ChatParticipant
	if err := h.db.Where("chat_id = ? AND user_id = ?", chatID, userID).First(&userPart).Error; err != nil {
		return fiber.ErrForbidden
	}

	// Verify chat is group chat
	var chat models.Chat
	if err := h.db.First(&chat, "id = ? AND is_group = true", chatID).Error; err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "not a group chat")
	}

	// Remove member
	if err := h.db.Where("chat_id = ? AND user_id = ?", chatID, removeMemberID).Delete(&models.ChatParticipant{}).Error; err != nil {
		return fiber.ErrInternalServerError
	}

	h.db.Preload("Participants.User.Profile").First(&chat, chat.ID)
	return c.JSON(chat)
}
