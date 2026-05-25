package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"student-marketplace/internal/middleware"
	"student-marketplace/internal/models"
)

type ServiceHandler struct {
	db *gorm.DB
}

func NewServiceHandler(db *gorm.DB) *ServiceHandler {
	return &ServiceHandler{db: db}
}

type CreateServiceRequest struct {
	CategoryID  string   `json:"category_id" validate:"required,uuid"`
	Title       string   `json:"title" validate:"required,min=10,max=200"`
	Description string   `json:"description" validate:"required,min=50"`
	Tags        []string `json:"tags"`
	Gallery     []string `json:"gallery"`
	Packages    []PackageInput `json:"packages" validate:"required,min=1"`
	FAQs        []FAQInput     `json:"faqs"`
}

type PackageInput struct {
	Name         string   `json:"name" validate:"required,oneof=basic standard premium"`
	Title        string   `json:"title" validate:"required"`
	Description  string   `json:"description"`
	Price        float64  `json:"price" validate:"required,min=1"`
	Currency     string   `json:"currency" validate:"required,oneof=USD KZT EUR"`
	DeliveryDays int      `json:"delivery_days" validate:"required,min=1"`
	Revisions    int      `json:"revisions" validate:"min=0"`
	Features     []string `json:"features"`
}

type FAQInput struct {
	Question string `json:"question" validate:"required"`
	Answer   string `json:"answer" validate:"required"`
}

// List / Search Services
func (h *ServiceHandler) List(c *fiber.Ctx) error {
	query := h.db.Model(&models.Service{}).
		Preload("Seller.Profile").
		Preload("Category").
		Preload("Packages").
		Where("is_active = true")

	// Search
	if search := c.Query("q"); search != "" {
		query = query.Where("to_tsvector('english', title || ' ' || description) @@ plainto_tsquery('english', ?)", search)
	}

	// Category filter
	if categorySlug := c.Query("category"); categorySlug != "" {
		var cat models.Category
		if h.db.Where("slug = ?", categorySlug).First(&cat).Error == nil {
			query = query.Where("category_id = ?", cat.ID)
		}
	}

	// Price filter
	if minPrice := c.QueryFloat("min_price", 0); minPrice > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM service_packages sp WHERE sp.service_id = services.id AND sp.price >= ?)", minPrice)
	}
	if maxPrice := c.QueryFloat("max_price", 0); maxPrice > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM service_packages sp WHERE sp.service_id = services.id AND sp.price <= ?)", maxPrice)
	}

	// Rating filter
	if minRating := c.QueryFloat("min_rating", 0); minRating > 0 {
		query = query.Where("rating >= ?", minRating)
	}

	// Sort
	switch c.Query("sort", "trending") {
	case "newest":
		query = query.Order("created_at DESC")
	case "price_asc":
		query = query.Order("(SELECT MIN(price) FROM service_packages sp WHERE sp.service_id = services.id) ASC")
	case "price_desc":
		query = query.Order("(SELECT MIN(price) FROM service_packages sp WHERE sp.service_id = services.id) DESC")
	case "rating":
		query = query.Order("rating DESC")
	default: // trending
		query = query.Order("is_featured DESC, orders_count DESC, rating DESC")
	}

	// Pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	if limit > 50 {
		limit = 50
	}
	offset := (page - 1) * limit

	var total int64
	query.Count(&total)

	var services []models.Service
	if err := query.Limit(limit).Offset(offset).Find(&services).Error; err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"data":  services,
		"total": total,
		"page":  page,
		"limit": limit,
		"pages": (int(total) + limit - 1) / limit,
	})
}

// Get Single Service
func (h *ServiceHandler) Get(c *fiber.Ctx) error {
	slugOrID := c.Params("slug")

	var service models.Service
	query := h.db.
		Preload("Seller.Profile").
		Preload("Category").
		Preload("Packages").
		Preload("FAQs").
		Preload("Reviews", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(10)
		}).
		Preload("Reviews.Reviewer.Profile")

	err := query.Where("slug = ? OR id::text = ?", slugOrID, slugOrID).First(&service).Error
	if err != nil {
		return fiber.ErrNotFound
	}

	// Increment views async
	go h.db.Model(&service).UpdateColumn("views", gorm.Expr("views + 1"))

	return c.JSON(service)
}

// Create Service
func (h *ServiceHandler) Create(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var req CreateServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	catID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid category_id")
	}

	slug := generateSlug(req.Title) + "-" + generateToken(4)

	service := models.Service{
		SellerID:    userID,
		CategoryID:  catID,
		Title:       req.Title,
		Slug:        slug,
		Description: req.Description,
		Tags:        req.Tags,
		Gallery:     req.Gallery,
	}

	tx := h.db.Begin()
	if err := tx.Create(&service).Error; err != nil {
		tx.Rollback()
		return fiber.ErrInternalServerError
	}

	for i, pkg := range req.Packages {
		p := models.ServicePackage{
			ServiceID:    service.ID,
			Name:         pkg.Name,
			Title:        pkg.Title,
			Description:  pkg.Description,
			Price:        pkg.Price,
			Currency:     pkg.Currency,
			DeliveryDays: pkg.DeliveryDays,
			Revisions:    pkg.Revisions,
			Features:     pkg.Features,
		}
		if err := tx.Create(&p).Error; err != nil {
			tx.Rollback()
			return fiber.ErrInternalServerError
		}
		_ = i
	}

	for i, faq := range req.FAQs {
		f := models.ServiceFAQ{
			ServiceID: service.ID,
			Question:  faq.Question,
			Answer:    faq.Answer,
			SortOrder: i,
		}
		tx.Create(&f)
	}

	// Promote seller role
	h.db.Model(&models.User{}).Where("id = ? AND role = ?", userID, models.RoleBuyer).
		Update("role", models.RoleSeller)

	tx.Commit()

	h.db.Preload("Packages").Preload("FAQs").Preload("Category").First(&service, service.ID)

	return c.Status(fiber.StatusCreated).JSON(service)
}

// Update Service
func (h *ServiceHandler) Update(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	serviceID := c.Params("id")
	var service models.Service
	if err := h.db.Where("id = ? AND seller_id = ?", serviceID, userID).First(&service).Error; err != nil {
		return fiber.ErrNotFound
	}

	var req CreateServiceRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	updates := map[string]any{
		"title":       req.Title,
		"description": req.Description,
		"tags":        req.Tags,
		"gallery":     req.Gallery,
		"updated_at":  time.Now(),
	}
	h.db.Model(&service).Updates(updates)

	return c.JSON(service)
}

// Delete Service
func (h *ServiceHandler) Delete(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	serviceID := c.Params("id")
	result := h.db.Where("id = ? AND seller_id = ?", serviceID, userID).Delete(&models.Service{})
	if result.RowsAffected == 0 {
		return fiber.ErrNotFound
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Featured Services
func (h *ServiceHandler) Featured(c *fiber.Ctx) error {
	var services []models.Service
	h.db.
		Preload("Seller.Profile").
		Preload("Category").
		Preload("Packages").
		Where("is_active = true AND is_featured = true").
		Order("orders_count DESC").
		Limit(8).
		Find(&services)

	return c.JSON(services)
}

// Categories
func (h *ServiceHandler) Categories(c *fiber.Ctx) error {
	var categories []models.Category
	h.db.Where("parent_id IS NULL").Order("sort_order").Find(&categories)
	return c.JSON(categories)
}

// MyServices — current user's own services
func (h *ServiceHandler) MyServices(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	var services []models.Service
	h.db.Preload("Packages").Preload("Category").
		Where("seller_id = ?", userID).
		Order("created_at DESC").
		Find(&services)
	return c.JSON(services)
}

// ─── HELPERS ─────────────────────────────────────────────────

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	s := result.String()
	if len(s) > 100 {
		s = s[:100]
	}
	return fmt.Sprintf("%s-%d", s, time.Now().UnixMilli()%10000)
}
