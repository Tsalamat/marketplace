package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"student-marketplace/internal/middleware"
	"student-marketplace/internal/models"
)

type PostHandler struct {
	db *gorm.DB
}

func NewPostHandler(db *gorm.DB) *PostHandler {
	return &PostHandler{db: db}
}

type CreatePostRequest struct {
	Content string   `json:"content" validate:"required,min=1,max=5000"`
	Images  []string `json:"images"`
}

// Community Feed
func (h *PostHandler) Feed(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	offset := (page - 1) * limit

	var posts []models.Post
	h.db.
		Preload("Author.Profile").
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(2)
		}).
		Preload("Comments.Author.Profile").
		Order("is_pinned DESC, created_at DESC").
		Limit(limit).Offset(offset).
		Find(&posts)

	// Check if liked by current user
	emptyID := uuid.UUID{}
	if userID != emptyID {
		for i := range posts {
			var likeCount int64
			h.db.Model(&models.Like{}).
				Where("user_id = ? AND post_id = ?", userID, posts[i].ID).
				Count(&likeCount)
			posts[i].IsLiked = likeCount > 0
		}
	}

	return c.JSON(posts)
}

// Create Post
func (h *PostHandler) Create(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var req CreatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	post := models.Post{
		AuthorID: userID,
		Content:  req.Content,
		Images:   req.Images,
	}
	h.db.Create(&post)
	h.db.Preload("Author.Profile").First(&post, post.ID)

	return c.Status(fiber.StatusCreated).JSON(post)
}

// Like / Unlike Post
func (h *PostHandler) ToggleLike(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	postID := c.Params("id")

	var existing models.Like
	err := h.db.Where("user_id = ? AND post_id = ?", userID, postID).First(&existing).Error

	if err == nil {
		// Unlike
		h.db.Delete(&existing)
		h.db.Model(&models.Post{}).Where("id = ?", postID).
			UpdateColumn("likes_count", gorm.Expr("likes_count - 1"))
		return c.JSON(fiber.Map{"liked": false})
	}

	// Like
	postUUID := mustParseUUID(postID)
	h.db.Create(&models.Like{UserID: userID, PostID: &postUUID})
	h.db.Model(&models.Post{}).Where("id = ?", postID).
		UpdateColumn("likes_count", gorm.Expr("likes_count + 1"))

	return c.JSON(fiber.Map{"liked": true})
}

// GET /api/v1/posts/:id/comments
func (h *PostHandler) GetComments(c *fiber.Ctx) error {
	postID := c.Params("id")
	var comments []models.Comment
	h.db.Preload("Author.Profile").
		Where("post_id = ?", postID).
		Order("created_at ASC").
		Find(&comments)
	return c.JSON(comments)
}

// Add Comment
func (h *PostHandler) AddComment(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	postID := c.Params("id")

	var body struct {
		Content  string  `json:"content" validate:"required"`
		ParentID *string `json:"parent_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	comment := models.Comment{
		PostID:   mustParseUUID(postID),
		AuthorID: userID,
		Content:  body.Content,
	}
	if body.ParentID != nil && *body.ParentID != "" {
		parentUUID := mustParseUUID(*body.ParentID)
		comment.ParentID = &parentUUID
	}

	h.db.Create(&comment)
	h.db.Model(&models.Post{}).Where("id = ?", postID).
		UpdateColumn("comments_count", gorm.Expr("comments_count + 1"))
	h.db.Preload("Author.Profile").First(&comment, comment.ID)

	return c.Status(fiber.StatusCreated).JSON(comment)
}

// Follow User
func (h *PostHandler) Follow(c *fiber.Ctx) error {
	followerID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	followingID := mustParseUUID(c.Params("userId"))
	if followerID == followingID {
		return fiber.NewError(fiber.StatusBadRequest, "cannot follow yourself")
	}

	var existing models.Follow
	err := h.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&existing).Error

	if err == nil {
		// Unfollow
		h.db.Delete(&existing)
		return c.JSON(fiber.Map{"following": false})
	}

	h.db.Create(&models.Follow{FollowerID: followerID, FollowingID: followingID})
	return c.JSON(fiber.Map{"following": true})
}

// Send Friend Request
func (h *PostHandler) SendFriendRequest(c *fiber.Ctx) error {
	requesterID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	addresseeID := mustParseUUID(c.Params("userId"))
	if requesterID == addresseeID {
		return fiber.NewError(fiber.StatusBadRequest, "cannot friend yourself")
	}

	var existing models.Friendship
	if h.db.Where("(requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)",
		requesterID, addresseeID, addresseeID, requesterID).First(&existing).Error == nil {
		return fiber.NewError(fiber.StatusConflict, "friendship request already exists")
	}

	friendship := models.Friendship{
		RequesterID: requesterID,
		AddresseeID: addresseeID,
		Status:      models.FriendPending,
	}
	h.db.Create(&friendship)

	return c.Status(fiber.StatusCreated).JSON(friendship)
}

// Accept / Reject Friend Request
func (h *PostHandler) RespondFriendRequest(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	friendshipID := c.Params("id")
	var friendship models.Friendship
	if err := h.db.Where("id = ? AND addressee_id = ?", friendshipID, userID).First(&friendship).Error; err != nil {
		return fiber.ErrNotFound
	}

	var body struct {
		Action string `json:"action"` // accept, reject
	}
	c.BodyParser(&body)

	switch body.Action {
	case "accept":
		h.db.Model(&friendship).Update("status", models.FriendAccepted)
	case "reject":
		h.db.Delete(&friendship)
	default:
		return fiber.NewError(fiber.StatusBadRequest, "action must be accept or reject")
	}

	return c.JSON(fiber.Map{"message": "ok"})
}

// Delete Post
func (h *PostHandler) Delete(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	postID := c.Params("id")
	role := middleware.GetUserRole(c)
	query := h.db.Where("id = ?", postID)
	if role != string(models.RoleAdmin) {
		query = query.Where("author_id = ?", userID)
	}
	if result := query.Delete(&models.Post{}); result.RowsAffected == 0 {
		return fiber.ErrNotFound
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// UUID alias for models package
type UUID = string
