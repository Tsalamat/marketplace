package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"student-marketplace/internal/middleware"
	"student-marketplace/internal/models"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// GET /api/v1/users/:username
func (h *UserHandler) GetByUsername(c *fiber.Ctx) error {
	username := c.Params("username")

	var user models.User
	if err := h.db.Preload("Profile").Where("username = ?", username).First(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	// Load user's services
	var services []models.Service
	h.db.Preload("Packages").Preload("Category").
		Where("seller_id = ? AND is_active = true", user.ID).
		Order("orders_count DESC").
		Limit(12).
		Find(&services)

	return c.JSON(fiber.Map{
		"id":            user.ID,
		"email":         user.Email,
		"username":      user.Username,
		"role":          user.Role,
		"email_verified": user.EmailVerified,
		"last_active":   user.LastActive,
		"created_at":    user.CreatedAt,
		"profile":       user.Profile,
		"services":      services,
	})
}

// GET /api/v1/users/search?q=...&limit=20
func (h *UserHandler) Search(c *fiber.Ctx) error {
	q := c.Query("q", "")
	limit := c.QueryInt("limit", 20)

	var users []models.User
	query := h.db.Preload("Profile").Limit(limit)
	if q != "" {
		like := "%" + q + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ?", like, like)
	}
	query.Order("created_at DESC").Find(&users)

	type Row struct {
		ID       string         `json:"id"`
		Username string         `json:"username"`
		Profile  *models.Profile `json:"profile"`
	}
	result := make([]Row, 0, len(users))
	for _, u := range users {
		result = append(result, Row{ID: u.ID.String(), Username: u.Username, Profile: u.Profile})
	}
	return c.JSON(result)
}

// GET /api/v1/users/me
func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var user models.User
	if err := h.db.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}
	return c.JSON(fiber.Map{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
		"role":     user.Role,
		"profile":  user.Profile,
	})
}

// PUT /api/v1/users/profile
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var body struct {
		FirstName   string   `json:"first_name"`
		LastName    string   `json:"last_name"`
		Tagline     string   `json:"tagline"`
		Bio         string   `json:"bio"`
		Skills      []string `json:"skills"`
		Languages   []string `json:"languages"`
		Location    string   `json:"location"`
		University  string   `json:"university"`
		Department  string   `json:"department"`
		YearOfStudy int      `json:"year_of_study"`
		AvatarURL   string   `json:"avatar_url"`
		CoverURL    string   `json:"cover_url"`
		GithubURL   string   `json:"github_url"`
		LinkedinURL string   `json:"linkedin_url"`
		PortfolioURL string  `json:"portfolio_url"`
		CurrencyPref string  `json:"currency_pref"`
	}

	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	updates := map[string]any{
		"first_name":    body.FirstName,
		"last_name":     body.LastName,
		"tagline":       body.Tagline,
		"bio":           body.Bio,
		"location":      body.Location,
		"university":    body.University,
		"department":    body.Department,
		"github_url":    body.GithubURL,
		"linkedin_url":  body.LinkedinURL,
		"portfolio_url": body.PortfolioURL,
	}
	if len(body.Skills) > 0 {
		updates["skills"] = body.Skills
	}
	if len(body.Languages) > 0 {
		updates["languages"] = body.Languages
	}
	if body.YearOfStudy > 0 {
		updates["year_of_study"] = body.YearOfStudy
	}
	if body.AvatarURL != "" {
		updates["avatar_url"] = body.AvatarURL
	}
	if body.CoverURL != "" {
		updates["cover_url"] = body.CoverURL
	}
	if body.CurrencyPref != "" {
		updates["currency_pref"] = body.CurrencyPref
	}

	result := h.db.Model(&models.Profile{}).Where("user_id = ?", userID).Updates(updates)
	if result.Error != nil {
		return fiber.ErrInternalServerError
	}

	var profile models.Profile
	h.db.Where("user_id = ?", userID).First(&profile)
	return c.JSON(profile)
}

// PATCH /api/v1/users/location
func (h *UserHandler) UpdateLocation(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var body struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	now := time.Now()
	h.db.Model(&models.Profile{}).Where("user_id = ?", userID).Updates(map[string]any{
		"lat":         body.Lat,
		"lng":         body.Lng,
		"location_at": now,
	})
	return c.JSON(fiber.Map{"lat": body.Lat, "lng": body.Lng})
}
