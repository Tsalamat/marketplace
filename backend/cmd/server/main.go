package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	miniogo "github.com/minio/minio-go/v7"
	miniocreds "github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"student-marketplace/internal/config"
	"student-marketplace/internal/handlers"
	"student-marketplace/internal/metrics"
	"student-marketplace/internal/middleware"
	"student-marketplace/internal/models"
	wsHub "student-marketplace/internal/websocket"
)

func main() {
	cfg := config.Load()

	// ─── Database ─────────────────────────────────────────────
	db, err := gorm.Open(postgres.Open(cfg.DB.DSN), &gorm.Config{
		Logger: gormlogger.Default.LogMode(func() gormlogger.LogLevel {
			if cfg.App.Env == "production" {
				return gormlogger.Error
			}
			return gormlogger.Warn
		}()),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	if err := db.AutoMigrate(
		&models.User{},
		&models.Profile{},
		&models.Category{},
		&models.Service{},
		&models.ServicePackage{},
		&models.ServiceFAQ{},
		&models.Order{},
		&models.Chat{},
		&models.ChatParticipant{},
		&models.ChatMessage{},
		&models.Post{},
		&models.Comment{},
		&models.Like{},
		&models.Follow{},
		&models.Friendship{},
		&models.Notification{},
		&models.Review{},
		&models.Report{},
		&models.AdminLog{},
	); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	seedCategories(db)

	// ─── MinIO ────────────────────────────────────────────────
	mc, err := miniogo.New(cfg.MinIO.Endpoint, &miniogo.Options{
		Creds:  miniocreds.NewStaticV4(cfg.MinIO.User, cfg.MinIO.Password, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		log.Printf("WARNING: MinIO unavailable (%v) — uploads disabled", err)
		mc = nil
	} else {
		// ensure bucket exists
		ctx := context.Background()
		exists, _ := mc.BucketExists(ctx, cfg.MinIO.Bucket)
		if !exists {
			if err := mc.MakeBucket(ctx, cfg.MinIO.Bucket, miniogo.MakeBucketOptions{}); err != nil {
				log.Printf("WARNING: could not create bucket: %v", err)
			} else {
				// make bucket public-read
				policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::` + cfg.MinIO.Bucket + `/*"]}]}`
				_ = mc.SetBucketPolicy(ctx, cfg.MinIO.Bucket, policy)
				log.Printf("Created bucket: %s", cfg.MinIO.Bucket)
			}
		}
	}

	// ─── WebSocket Hub ────────────────────────────────────────
	hub := wsHub.NewHub(db)
	go hub.Run()

	// ─── Fiber ───────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "Student Marketplace API v1",
		ErrorHandler: errorHandler,
		BodyLimit:    50 * 1024 * 1024, // 50 MB for file uploads
	})

	app.Use(recover.New())
	app.Use(metrics.Middleware())
	app.Use(logger.New(logger.Config{Format: "[${time}] ${status} ${method} ${path} ${latency}\n"}))
	app.Use(helmet.New())
	app.Use(compress.New())

	frontendOrigin := cfg.App.FrontendURL
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:3000"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     frontendOrigin,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// ─── Handlers ─────────────────────────────────────────────
	authH    := handlers.NewAuthHandler(db)
	oauthH   := handlers.NewOAuthHandler(db)
	userH    := handlers.NewUserHandler(db)
	serviceH := handlers.NewServiceHandler(db)
	orderH   := handlers.NewOrderHandler(db)
	chatH    := handlers.NewChatHandler(db)
	postH    := handlers.NewPostHandler(db)
	notifH   := handlers.NewNotificationHandler(db)
	adminH   := handlers.NewAdminHandler(db)
	friendH  := handlers.NewFriendHandler(db)
	uploadH  := handlers.NewUploadHandler(mc, cfg)

	api := app.Group("/api/v1")

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "version": "1.0.0"})
	})
	app.Get("/metrics", metrics.Handler)

	// ── Auth ──────────────────────────────────────────────────
	auth := api.Group("/auth")
	auth.Post("/register",        authH.Register)
	auth.Post("/login",           authH.Login)
	auth.Post("/refresh",         authH.Refresh)
	auth.Post("/logout",          authH.Logout)
	auth.Get("/verify-email",     authH.VerifyEmail)
	auth.Post("/forgot-password", authH.ForgotPassword)
	auth.Post("/reset-password",  authH.ResetPassword)
	auth.Post("/change-password", middleware.RequireAuth, authH.ChangePassword)
	auth.Get("/me",               middleware.RequireAuth, authH.Me)
	auth.Get("/google",            oauthH.Redirect)
	auth.Get("/google/callback",   oauthH.Callback)
	auth.Post("/google/id-token",  oauthH.VerifyIDToken) // native mobile sign-in
	auth.Get("/oauth-tokens",      oauthH.GetTokens)

	// ── Upload ────────────────────────────────────────────────
	upload := api.Group("/upload", middleware.RequireAuth)
	upload.Post("/image",    uploadH.UploadImage)
	upload.Post("/file",     uploadH.UploadFile)
	upload.Post("/avatar",   uploadH.UploadAvatar)

	// ── Users ─────────────────────────────────────────────────
	users := api.Group("/users")
	users.Get("/search",     middleware.OptionalAuth, userH.Search)
	users.Get("/me",         middleware.RequireAuth, userH.GetMe)
	users.Put("/profile",    middleware.RequireAuth, userH.UpdateProfile)
	users.Patch("/profile",  middleware.RequireAuth, userH.UpdateProfile)
	users.Patch("/location", middleware.RequireAuth, userH.UpdateLocation)
	users.Get("/:username",  userH.GetByUsername)

	// ── Services ──────────────────────────────────────────────
	services := api.Group("/services")
	services.Get("/",           middleware.OptionalAuth, serviceH.List)
	services.Get("/featured",   serviceH.Featured)
	services.Get("/categories", serviceH.Categories)
	services.Get("/my",         middleware.RequireAuth, serviceH.MyServices)
	services.Get("/:slug",      middleware.OptionalAuth, serviceH.Get)
	services.Post("/",          middleware.RequireAuth, serviceH.Create)
	services.Put("/:id",        middleware.RequireAuth, serviceH.Update)
	services.Delete("/:id",     middleware.RequireAuth, serviceH.Delete)

	// ── Orders ────────────────────────────────────────────────
	orders := api.Group("/orders", middleware.RequireAuth)
	orders.Post("/",             orderH.Create)
	orders.Get("/",              orderH.List)
	orders.Get("/:id",           orderH.Get)
	orders.Patch("/:id/status",  orderH.UpdateStatus)
	orders.Post("/:id/delivery", orderH.SubmitDelivery)
	orders.Post("/:id/review",   orderH.LeaveReview)

	// ── Chat ──────────────────────────────────────────────────
	chat := api.Group("/chat", middleware.RequireAuth)
	chat.Get("/",                                chatH.ListChats)
	chat.Get("/:chatId/messages",                chatH.GetMessages)
	chat.Post("/direct/:userId",                 chatH.GetOrCreateDirect)
	chat.Post("/group",                          chatH.CreateGroupChat)
	chat.Post("/:chatId/members/:userId",        chatH.AddMember)
	chat.Delete("/:chatId/members/:userId",      chatH.RemoveMember)

	// ── Friends ───────────────────────────────────────────────
	friends := api.Group("/friends", middleware.RequireAuth)
	friends.Get("/",           friendH.List)
	friends.Get("/requests",   friendH.Requests)
	friends.Post("/:userId",   friendH.Send)
	friends.Put("/:id/accept", friendH.Accept)
	friends.Put("/:id/reject", friendH.Reject)
	friends.Delete("/:userId", friendH.Remove)
	friends.Get("/locations",  friendH.Locations)

	// ── Posts ─────────────────────────────────────────────────
	posts := api.Group("/posts")
	posts.Get("/",              middleware.OptionalAuth, postH.Feed)
	posts.Post("/",             middleware.RequireAuth,  postH.Create)
	posts.Post("/:id/like",     middleware.RequireAuth,  postH.ToggleLike)
	posts.Get("/:id/comments",  middleware.OptionalAuth, postH.GetComments)
	posts.Post("/:id/comments", middleware.RequireAuth,  postH.AddComment)
	posts.Delete("/:id",        middleware.RequireAuth,  postH.Delete)

	// ── Social (follow) ───────────────────────────────────────
	social := api.Group("/social", middleware.RequireAuth)
	social.Post("/follow/:userId", postH.Follow)

	// ── Notifications ─────────────────────────────────────────
	notifs := api.Group("/notifications", middleware.RequireAuth)
	notifs.Get("/",            notifH.List)
	notifs.Patch("/:id/read",  notifH.MarkRead)
	notifs.Post("/read-all",   notifH.MarkAllRead)
	notifs.Delete("/:id",      notifH.Delete)

	// ── Admin ─────────────────────────────────────────────────
	admin := api.Group("/admin",
		middleware.RequireAuth,
		middleware.RequireRole(models.RoleAdmin))
	admin.Get("/dashboard",          adminH.Dashboard)
	admin.Get("/users",              adminH.ListUsers)
	admin.Post("/users/:id/ban",     adminH.BanUser)
	admin.Post("/users/:id/unban",   adminH.UnbanUser)
	admin.Patch("/services/:id",     adminH.ModerateService)
	admin.Get("/reports",            adminH.ListReports)
	admin.Post("/reports/:id/resolve", adminH.ResolveReport)
	admin.Get("/revenue",            adminH.Revenue)
	admin.Post("/seed",              adminH.Seed)

	// ── WebSocket (chat + WebRTC signaling) ───────────────────
	app.Get("/ws/chat", wsHub.WSUpgradeMiddleware, hub.Handler())

	addr := fmt.Sprintf(":%s", cfg.App.Port)
	log.Printf("Student Marketplace API → %s  [%s]", addr, cfg.App.Env)
	log.Fatal(app.Listen(addr))
}

func seedCategories(db *gorm.DB) {
	var count int64
	db.Model(&models.Category{}).Count(&count)
	if count > 0 { return }
	cats := []models.Category{
		{Name: "Programming & Tech",    Slug: "programming",  Icon: "code",       SortOrder: 1},
		{Name: "Design & Creative",     Slug: "design",       Icon: "palette",    SortOrder: 2},
		{Name: "Tutoring & Education",  Slug: "tutoring",     Icon: "book-open",  SortOrder: 3},
		{Name: "Writing & Translation", Slug: "writing",      Icon: "pen-tool",   SortOrder: 4},
		{Name: "Video & Animation",     Slug: "video",        Icon: "video",      SortOrder: 5},
		{Name: "Photography",           Slug: "photography",  Icon: "camera",     SortOrder: 6},
		{Name: "Delivery & Errands",    Slug: "delivery",     Icon: "truck",      SortOrder: 7},
		{Name: "Fitness & Wellness",    Slug: "fitness",      Icon: "activity",   SortOrder: 8},
		{Name: "Music & Audio",         Slug: "music",        Icon: "music",      SortOrder: 9},
		{Name: "Business",              Slug: "business",     Icon: "briefcase",  SortOrder: 10},
	}
	db.Create(&cats)
	log.Println("Categories seeded")
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	msg  := "internal server error"
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		msg  = e.Message
	}
	return c.Status(code).JSON(fiber.Map{"error": msg, "status": code})
}
