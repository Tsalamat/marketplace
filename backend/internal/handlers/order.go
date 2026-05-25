package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"student-marketplace/internal/middleware"
	"student-marketplace/internal/models"
	"student-marketplace/internal/config"
)

type OrderHandler struct {
	db *gorm.DB
}

func NewOrderHandler(db *gorm.DB) *OrderHandler {
	return &OrderHandler{db: db}
}

type CreateOrderRequest struct {
	ServiceID    string `json:"service_id" validate:"required,uuid"`
	PackageID    string `json:"package_id" validate:"required,uuid"`
	Requirements string `json:"requirements"`
}

type ReviewRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Content string `json:"content"`
}

// Create Order (Place Order)
func (h *OrderHandler) Create(c *fiber.Ctx) error {
	buyerID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	serviceID, err := uuid.Parse(req.ServiceID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid service_id")
	}
	packageID, err := uuid.Parse(req.PackageID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid package_id")
	}

	// Load service & package
	var service models.Service
	if err := h.db.First(&service, "id = ? AND is_active = true", serviceID).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "service not found")
	}

	if service.SellerID == buyerID {
		return fiber.NewError(fiber.StatusBadRequest, "cannot order your own service")
	}

	var pkg models.ServicePackage
	if err := h.db.First(&pkg, "id = ? AND service_id = ?", packageID, serviceID).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "package not found")
	}

	feePercent := config.Cfg.Platform.FeePercent / 100
	platformFee := pkg.Price * feePercent
	sellerAmount := pkg.Price - platformFee
	dueDate := time.Now().AddDate(0, 0, pkg.DeliveryDays)

	order := models.Order{
		BuyerID:      buyerID,
		SellerID:     service.SellerID,
		ServiceID:    serviceID,
		PackageID:    packageID,
		Status:       models.OrderPending,
		Amount:       pkg.Price,
		PlatformFee:  platformFee,
		SellerAmount: sellerAmount,
		Currency:     pkg.Currency,
		Requirements: req.Requirements,
		MaxRevisions: pkg.Revisions,
		DueDate:      &dueDate,
	}

	if err := h.db.Create(&order).Error; err != nil {
		return fiber.ErrInternalServerError
	}

	// Create chat for this order
	chat := models.Chat{OrderID: &order.ID}
	h.db.Create(&chat)
	h.db.Create(&models.ChatParticipant{ChatID: chat.ID, UserID: buyerID})
	h.db.Create(&models.ChatParticipant{ChatID: chat.ID, UserID: service.SellerID})

	// Update service stats
	h.db.Model(&service).UpdateColumn("orders_count", gorm.Expr("orders_count + 1"))

	// Notify seller
	h.createNotification(service.SellerID, models.NotifOrder,
		"New Order Received!",
		"You have a new order for: "+service.Title,
		map[string]any{"order_id": order.ID})

	h.db.Preload("Buyer.Profile").Preload("Seller.Profile").
		Preload("Service").Preload("Package").First(&order, order.ID)

	return c.Status(fiber.StatusCreated).JSON(order)
}

// List Orders (for current user)
func (h *OrderHandler) List(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	as := c.Query("as", "buyer") // buyer or seller

	query := h.db.Model(&models.Order{}).
		Preload("Buyer.Profile").
		Preload("Seller.Profile").
		Preload("Service").
		Preload("Package")

	if as == "seller" {
		query = query.Where("seller_id = ?", userID)
	} else {
		query = query.Where("buyer_id = ?", userID)
	}

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var orders []models.Order
	query.Order("created_at DESC").Find(&orders)

	return c.JSON(orders)
}

// Get Single Order
func (h *OrderHandler) Get(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	orderID := c.Params("id")
	var order models.Order
	err := h.db.
		Preload("Buyer.Profile").
		Preload("Seller.Profile").
		Preload("Service").
		Preload("Package").
		Preload("Review").
		Where("id = ? AND (buyer_id = ? OR seller_id = ?)", orderID, userID, userID).
		First(&order).Error

	if err != nil {
		return fiber.ErrNotFound
	}

	return c.JSON(order)
}

// Update Order Status
func (h *OrderHandler) UpdateStatus(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	orderID := c.Params("id")
	var order models.Order
	if err := h.db.Where("id = ? AND (buyer_id = ? OR seller_id = ?)", orderID, userID, userID).
		First(&order).Error; err != nil {
		return fiber.ErrNotFound
	}

	var body struct {
		Status models.OrderStatus `json:"status"`
		Reason string             `json:"reason"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	// Validate transitions
	switch body.Status {
	case models.OrderInProgress:
		if order.Status != models.OrderPending {
			return fiber.NewError(fiber.StatusBadRequest, "cannot transition to in_progress")
		}
		if order.SellerID != userID {
			return fiber.NewError(fiber.StatusForbidden, "only seller can accept order")
		}
	case models.OrderDelivered:
		if order.Status != models.OrderInProgress {
			return fiber.NewError(fiber.StatusBadRequest, "cannot deliver order not in progress")
		}
		if order.SellerID != userID {
			return fiber.NewError(fiber.StatusForbidden, "only seller can deliver")
		}
		now := time.Now()
		h.db.Model(&order).Update("delivered_at", now)
	case models.OrderCompleted:
		if order.Status != models.OrderDelivered {
			return fiber.NewError(fiber.StatusBadRequest, "order must be delivered first")
		}
		if order.BuyerID != userID {
			return fiber.NewError(fiber.StatusForbidden, "only buyer can complete order")
		}
		now := time.Now()
		h.db.Model(&order).Update("completed_at", now)
		// Update seller stats
		h.db.Model(&models.Profile{}).Where("user_id = ?", order.SellerID).
			UpdateColumn("completed_jobs", gorm.Expr("completed_jobs + 1"))
	case models.OrderRevision:
		if order.Status != models.OrderDelivered {
			return fiber.NewError(fiber.StatusBadRequest, "can only request revision on delivered order")
		}
		if order.BuyerID != userID {
			return fiber.NewError(fiber.StatusForbidden, "only buyer can request revision")
		}
		if order.RevisionCount >= order.MaxRevisions {
			return fiber.NewError(fiber.StatusBadRequest, "no revisions remaining")
		}
		h.db.Model(&order).UpdateColumn("revision_count", gorm.Expr("revision_count + 1"))
	case models.OrderCancelled:
		if order.Status == models.OrderCompleted || order.Status == models.OrderCancelled {
			return fiber.NewError(fiber.StatusBadRequest, "cannot cancel order in current state")
		}
		now := time.Now()
		h.db.Model(&order).Updates(map[string]any{
			"cancel_reason": body.Reason,
			"cancelled_at":  now,
		})
	}

	h.db.Model(&order).Update("status", body.Status)

	// Notify the other party
	notifyID := order.BuyerID
	if order.BuyerID == userID {
		notifyID = order.SellerID
	}
	h.createNotification(notifyID, models.NotifOrder,
		"Order Update",
		"Your order status changed to: "+string(body.Status),
		map[string]any{"order_id": order.ID})

	return c.JSON(fiber.Map{"message": "status updated", "status": body.Status})
}

// Submit Delivery
func (h *OrderHandler) SubmitDelivery(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	orderID := c.Params("id")
	var order models.Order
	if err := h.db.Where("id = ? AND seller_id = ?", orderID, userID).First(&order).Error; err != nil {
		return fiber.ErrNotFound
	}

	if order.Status != models.OrderInProgress {
		return fiber.NewError(fiber.StatusBadRequest, "order must be in progress to submit delivery")
	}

	var body struct {
		Files   []string `json:"files"`
		Message string   `json:"message"`
	}
	c.BodyParser(&body)

	now := time.Now()
	h.db.Model(&order).Updates(map[string]any{
		"status":         models.OrderDelivered,
		"delivery_files": body.Files,
		"delivered_at":   now,
	})

	h.createNotification(order.BuyerID, models.NotifOrder,
		"Delivery Submitted!",
		"Your order has been delivered. Please review.",
		map[string]any{"order_id": order.ID})

	return c.JSON(fiber.Map{"message": "delivery submitted"})
}

// Leave Review
func (h *OrderHandler) LeaveReview(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	orderID := c.Params("id")
	var order models.Order
	if err := h.db.Where("id = ? AND buyer_id = ? AND status = ?",
		orderID, userID, models.OrderCompleted).First(&order).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "completed order not found")
	}

	// Check no existing review
	var existing models.Review
	if h.db.Where("order_id = ?", order.ID).First(&existing).Error == nil {
		return fiber.NewError(fiber.StatusConflict, "review already submitted")
	}

	var req ReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	review := models.Review{
		OrderID:    order.ID,
		ServiceID:  order.ServiceID,
		ReviewerID: userID,
		SellerID:   order.SellerID,
		Rating:     req.Rating,
		Content:    req.Content,
	}
	h.db.Create(&review)

	// Recalculate seller & service rating
	go h.recalculateRatings(order.ServiceID, order.SellerID)

	h.createNotification(order.SellerID, models.NotifReview,
		"New Review Received!",
		"You received a new review.",
		map[string]any{"order_id": order.ID})

	return c.Status(fiber.StatusCreated).JSON(review)
}

// ─── HELPERS ─────────────────────────────────────────────────

func (h *OrderHandler) recalculateRatings(serviceID, sellerID uuid.UUID) {
	var avgServiceRating float64
	var countService int64
	h.db.Model(&models.Review{}).Where("service_id = ?", serviceID).
		Select("AVG(rating)").Row().Scan(&avgServiceRating)
	h.db.Model(&models.Review{}).Where("service_id = ?", serviceID).Count(&countService)

	h.db.Model(&models.Service{}).Where("id = ?", serviceID).Updates(map[string]any{
		"rating":        avgServiceRating,
		"total_reviews": countService,
	})

	var avgSellerRating float64
	var countSeller int64
	h.db.Model(&models.Review{}).Where("seller_id = ?", sellerID).
		Select("AVG(rating)").Row().Scan(&avgSellerRating)
	h.db.Model(&models.Review{}).Where("seller_id = ?", sellerID).Count(&countSeller)

	h.db.Model(&models.Profile{}).Where("user_id = ?", sellerID).Updates(map[string]any{
		"rating":        avgSellerRating,
		"total_reviews": countSeller,
	})
}

func (h *OrderHandler) createNotification(userID uuid.UUID, ntype models.NotificationType, title, body string, data map[string]any) {
	notif := models.Notification{
		UserID: userID,
		Type:   ntype,
		Title:  title,
		Body:   body,
		Data:   data,
	}
	h.db.Create(&notif)
}
