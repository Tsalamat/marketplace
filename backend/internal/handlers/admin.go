package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"student-marketplace/internal/middleware"
	"student-marketplace/internal/models"
)

type AdminHandler struct {
	db *gorm.DB
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// Dashboard Analytics
func (h *AdminHandler) Dashboard(c *fiber.Ctx) error {
	var stats struct {
		TotalUsers     int64   `json:"total_users"`
		TotalServices  int64   `json:"total_services"`
		TotalOrders    int64   `json:"total_orders"`
		CompletedOrders int64  `json:"completed_orders"`
		TotalRevenue   float64 `json:"total_revenue"`
		PlatformFees   float64 `json:"platform_fees"`
		ActiveUsers    int64   `json:"active_users"`
		TotalPosts     int64   `json:"total_posts"`
	}

	h.db.Model(&models.User{}).Count(&stats.TotalUsers)
	h.db.Model(&models.Service{}).Count(&stats.TotalServices)
	h.db.Model(&models.Order{}).Count(&stats.TotalOrders)
	h.db.Model(&models.Order{}).Where("status = ?", models.OrderCompleted).Count(&stats.CompletedOrders)
	h.db.Model(&models.Order{}).Where("status = ?", models.OrderCompleted).
		Select("COALESCE(SUM(amount), 0)").Row().Scan(&stats.TotalRevenue)
	h.db.Model(&models.Order{}).Where("status = ?", models.OrderCompleted).
		Select("COALESCE(SUM(platform_fee), 0)").Row().Scan(&stats.PlatformFees)
	h.db.Model(&models.Profile{}).Where("is_online = true").Count(&stats.ActiveUsers)
	h.db.Model(&models.Post{}).Count(&stats.TotalPosts)

	return c.JSON(stats)
}

// List Users
func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)
	offset := (page - 1) * limit

	query := h.db.Model(&models.User{}).Preload("Profile")

	if search := c.Query("q"); search != "" {
		query = query.Where("email ILIKE ? OR username ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}

	var total int64
	query.Count(&total)

	var users []models.User
	query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users)

	return c.JSON(fiber.Map{"data": users, "total": total})
}

// Ban User
func (h *AdminHandler) BanUser(c *fiber.Ctx) error {
	adminID, _ := middleware.GetUserID(c)
	userID := c.Params("id")

	var body struct {
		Reason string `json:"reason"`
	}
	c.BodyParser(&body)

	result := h.db.Model(&models.User{}).Where("id = ?", userID).Update("is_active", false)
	if result.RowsAffected == 0 {
		return fiber.ErrNotFound
	}

	h.db.Create(&models.AdminLog{
		AdminID:    adminID,
		Action:     "ban_user",
		TargetType: "user",
		Metadata:   map[string]any{"reason": body.Reason, "user_id": userID},
		IPAddress:  c.IP(),
	})

	return c.JSON(fiber.Map{"message": "user banned"})
}

// Unban User
func (h *AdminHandler) UnbanUser(c *fiber.Ctx) error {
	userID := c.Params("id")
	result := h.db.Model(&models.User{}).Where("id = ?", userID).Update("is_active", true)
	if result.RowsAffected == 0 {
		return fiber.ErrNotFound
	}
	return c.JSON(fiber.Map{"message": "user unbanned"})
}

// Moderate Service
func (h *AdminHandler) ModerateService(c *fiber.Ctx) error {
	serviceID := c.Params("id")
	var body struct {
		Active   *bool  `json:"is_active"`
		Featured *bool  `json:"is_featured"`
		Reason   string `json:"reason"`
	}
	c.BodyParser(&body)

	updates := map[string]any{}
	if body.Active != nil {
		updates["is_active"] = *body.Active
	}
	if body.Featured != nil {
		updates["is_featured"] = *body.Featured
	}

	h.db.Model(&models.Service{}).Where("id = ?", serviceID).Updates(updates)
	return c.JSON(fiber.Map{"message": "service moderated"})
}

// List Reports
func (h *AdminHandler) ListReports(c *fiber.Ctx) error {
	var reports []models.Report
	h.db.
		Preload("Reporter").
		Where("status = 'pending'").
		Order("created_at DESC").
		Find(&reports)
	return c.JSON(reports)
}

// Resolve Report
func (h *AdminHandler) ResolveReport(c *fiber.Ctx) error {
	reportID := c.Params("id")
	var body struct {
		Status string `json:"status"` // resolved, dismissed
		Note   string `json:"note"`
	}
	c.BodyParser(&body)

	h.db.Model(&models.Report{}).Where("id = ?", reportID).Updates(map[string]any{
		"status":     body.Status,
		"admin_note": body.Note,
	})
	return c.JSON(fiber.Map{"message": "report resolved"})
}

// Revenue Analytics
func (h *AdminHandler) Revenue(c *fiber.Ctx) error {
	var dailyRevenue []struct {
		Date     string  `json:"date"`
		Revenue  float64 `json:"revenue"`
		Platform float64 `json:"platform"`
		Orders   int64   `json:"orders"`
	}

	h.db.Raw(`
		SELECT
			DATE(created_at) as date,
			SUM(amount) as revenue,
			SUM(platform_fee) as platform,
			COUNT(*) as orders
		FROM orders
		WHERE status = 'completed'
		  AND created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`).Scan(&dailyRevenue)

	return c.JSON(dailyRevenue)
}

// POST /api/v1/admin/seed — create demo data (development only)
func (h *AdminHandler) Seed(c *fiber.Ctx) error {
	adminID, _ := middleware.GetUserID(c)

	type seedUser struct {
		email, username, first, last, university, tagline string
		role                                              models.UserRole
		skills                                            []string
	}
	users := []seedUser{
		{"alice@mit.edu", "alice_dev", "Alice", "Johnson", "MIT", "Full-stack developer & React expert", models.RoleSeller, []string{"React", "Go", "PostgreSQL"}},
		{"bob@stanford.edu", "bob_design", "Bob", "Chen", "Stanford", "UI/UX Designer & Figma specialist", models.RoleSeller, []string{"Figma", "Illustrator", "CSS"}},
		{"carol@harvard.edu", "carol_tutor", "Carol", "Williams", "Harvard", "Math & Physics tutor, 5 years exp", models.RoleSeller, []string{"Math", "Physics", "Python"}},
		{"dave@caltech.edu", "dave_video", "Dave", "Kim", "Caltech", "Video editor & motion graphics", models.RoleSeller, []string{"Premiere", "After Effects", "DaVinci"}},
		{"eve@columbia.edu", "eve_writer", "Eve", "Martinez", "Columbia", "Content writer & translator EN/RU/ES", models.RoleSeller, []string{"Writing", "Translation", "SEO"}},
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("Demo1234!"), bcrypt.DefaultCost)
	created := 0

	for _, u := range users {
		var existing models.User
		if h.db.Where("email = ?", u.email).First(&existing).Error == nil {
			continue
		}
		user := models.User{
			Email: u.email, Username: u.username,
			PasswordHash: string(hash), Role: u.role,
			EmailVerified: true, IsActive: true,
		}
		if err := h.db.Create(&user).Error; err != nil {
			continue
		}
		h.db.Create(&models.Profile{
			UserID: user.ID, FirstName: u.first, LastName: u.last,
			University: u.university, Tagline: u.tagline, Skills: u.skills,
			Rating: 4.5 + float64(created)*0.1, CompletedJobs: 10 + created*5,
		})
		created++
	}

	// Seed services for each seller
	type seedSvc struct {
		sellerEmail, catSlug, title, desc string
		price                             float64
	}
	services := []seedSvc{
		{"alice@mit.edu", "programming", "I will build your React web app", "Professional React/Next.js development with TypeScript, TailwindCSS, and modern best practices.", 150},
		{"alice@mit.edu", "programming", "Full-stack Go + Vue SaaS platform", "Complete SaaS application with Go backend, Vue 3 frontend, PostgreSQL, JWT auth.", 500},
		{"bob@stanford.edu", "design", "Modern logo & brand identity design", "Unique, memorable logo with full brand guidelines, colors, and typography.", 80},
		{"bob@stanford.edu", "design", "UI/UX design in Figma with prototypes", "Complete app design system with interactive prototype and design handoff.", 200},
		{"carol@harvard.edu", "tutoring", "Math tutoring — Calculus & Linear Algebra", "Clear explanations, practice problems, exam prep. All levels welcome.", 25},
		{"carol@harvard.edu", "tutoring", "Python programming for beginners", "Learn Python from scratch with real projects. 1-on-1 sessions.", 30},
		{"dave@caltech.edu", "video", "Professional video editing & color grading", "YouTube, TikTok, Instagram reels. Fast delivery, cinematic look.", 60},
		{"eve@columbia.edu", "writing", "SEO blog articles & content writing", "Engaging, research-backed articles optimized for Google rankings.", 20},
		{"eve@columbia.edu", "writing", "Translation EN ↔ RU ↔ ES", "Native-quality translation for documents, websites, and marketing materials.", 15},
	}

	for _, s := range services {
		var seller models.User
		if h.db.Where("email = ?", s.sellerEmail).First(&seller).Error != nil {
			continue
		}
		var cat models.Category
		if h.db.Where("slug = ?", s.catSlug).First(&cat).Error != nil {
			continue
		}
		slug := fmt.Sprintf("%s-%d", s.catSlug, created)
		created++
		svc := models.Service{
			SellerID: seller.ID, CategoryID: cat.ID,
			Title: s.title, Slug: slug, Description: s.desc,
			IsActive: true, Rating: 4.7, TotalReviews: 8, OrdersCount: 15,
		}
		if err := h.db.Create(&svc).Error; err != nil { continue }
		h.db.Create(&models.ServicePackage{
			ServiceID: svc.ID, Name: "basic", Title: "Basic",
			Price: s.price, Currency: "USD", DeliveryDays: 3, Revisions: 2,
		})
		h.db.Create(&models.ServicePackage{
			ServiceID: svc.ID, Name: "standard", Title: "Standard",
			Price: s.price * 1.8, Currency: "USD", DeliveryDays: 5, Revisions: 5,
		})
		h.db.Create(&models.ServicePackage{
			ServiceID: svc.ID, Name: "premium", Title: "Premium",
			Price: s.price * 3, Currency: "USD", DeliveryDays: 7, Revisions: -1,
		})
	}

	// Seed community posts
	posts := []string{
		"Just finished a React dashboard for a client — TypeScript + TailwindCSS + Recharts. Super happy with how clean the code came out!",
		"Looking for a math tutor for Linear Algebra. DM me if you're available for weekly sessions!",
		"Tip for sellers: always over-deliver on the first order. Your rating is everything on this platform.",
		"New to the platform! I'm a graphic designer from Stanford. Check out my portfolio and let me know if you need any design work.",
		"Anyone else think audio calls in chat would be super useful? Just had a client who couldn't type fast enough 😄",
	}
	for i, content := range posts {
		var users2 []models.User
		h.db.Limit(5).Find(&users2)
		if i >= len(users2) { break }
		h.db.Create(&models.Post{AuthorID: users2[i].ID, Content: content})
	}

	h.db.Create(&models.AdminLog{
		AdminID: adminID, Action: "seed_data",
		Metadata: map[string]any{"created": created},
	})

	return c.JSON(fiber.Map{"message": "demo data seeded", "users_created": len(users), "services_created": len(services)})
}
