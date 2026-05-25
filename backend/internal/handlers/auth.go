package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"student-marketplace/internal/middleware"
	"student-marketplace/internal/models"
	jwtpkg "student-marketplace/pkg/jwt"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// ─── REQUEST / RESPONSE TYPES ────────────────────────────────

type RegisterRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Username   string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Password   string `json:"password" validate:"required,min=8"`
	FirstName  string `json:"first_name" validate:"required"`
	LastName   string `json:"last_name" validate:"required"`
	University string `json:"university"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

type AuthResponse struct {
	User         *models.User      `json:"user"`
	*jwtpkg.TokenPair
}

// ─── REGISTER ────────────────────────────────────────────────

// @Summary Register new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "Registration data"
// @Success 201 {object} AuthResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// Check duplicate email
	var existing models.User
	if err := h.db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existing).Error; err == nil {
		return fiber.NewError(fiber.StatusConflict, "email or username already taken")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	verifyToken := generateToken(32)
	user := models.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hash),
		Role:         models.RoleBuyer,
		VerifyToken:  &verifyToken,
	}

	if err := h.db.Create(&user).Error; err != nil {
		return fiber.ErrInternalServerError
	}

	// Create profile
	profile := models.Profile{
		UserID:     user.ID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		University: req.University,
	}
	h.db.Create(&profile)

	// TODO: send verification email

	tokens, err := jwtpkg.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return fiber.ErrInternalServerError
	}

	user.Profile = &profile
	return c.Status(fiber.StatusCreated).JSON(AuthResponse{User: &user, TokenPair: tokens})
}

// ─── LOGIN ───────────────────────────────────────────────────

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	var user models.User
	if err := h.db.Preload("Profile").Where("email = ?", req.Email).First(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}

	if !user.IsActive {
		return fiber.NewError(fiber.StatusForbidden, "account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}

	now := time.Now()
	h.db.Model(&user).Update("last_active", now)

	tokens, err := jwtpkg.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(AuthResponse{User: &user, TokenPair: tokens})
}

// ─── REFRESH ─────────────────────────────────────────────────

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	claims, err := jwtpkg.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token")
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", claims.UserID).Error; err != nil {
		return fiber.ErrUnauthorized
	}

	if !user.IsActive {
		return fiber.NewError(fiber.StatusForbidden, "account is deactivated")
	}

	tokens, err := jwtpkg.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(tokens)
}

// ─── LOGOUT ──────────────────────────────────────────────────

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// In production, add refresh token to Redis blacklist here
	c.ClearCookie("access_token", "refresh_token")
	return c.JSON(fiber.Map{"message": "logged out successfully"})
}

// ─── ME ──────────────────────────────────────────────────────

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var user models.User
	if err := h.db.Preload("Profile").First(&user, "id = ?", userID).Error; err != nil {
		return fiber.ErrNotFound
	}

	return c.JSON(user)
}

// ─── VERIFY EMAIL ────────────────────────────────────────────

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "token required")
	}

	result := h.db.Model(&models.User{}).
		Where("verify_token = ?", token).
		Updates(map[string]any{
			"email_verified": true,
			"verify_token":   nil,
		})

	if result.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid or expired token")
	}

	return c.JSON(fiber.Map{"message": "email verified successfully"})
}

// ─── FORGOT PASSWORD ─────────────────────────────────────────

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Don't reveal if email exists
		return c.JSON(fiber.Map{"message": "if the email exists, a reset link has been sent"})
	}

	resetToken := generateToken(32)
	expires := time.Now().Add(1 * time.Hour)

	h.db.Model(&user).Updates(map[string]any{
		"reset_token":   resetToken,
		"reset_expires": expires,
	})

	// TODO: send reset email

	return c.JSON(fiber.Map{"message": "if the email exists, a reset link has been sent"})
}

// ─── RESET PASSWORD ──────────────────────────────────────────

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	var user models.User
	if err := h.db.Where("reset_token = ? AND reset_expires > ?", req.Token, time.Now()).First(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid or expired reset token")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	h.db.Model(&user).Updates(map[string]any{
		"password_hash": string(hash),
		"reset_token":   nil,
		"reset_expires": nil,
	})

	return c.JSON(fiber.Map{"message": "password reset successfully"})
}

// ─── CHANGE PASSWORD ─────────────────────────────────────────

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	var user models.User
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		return fiber.ErrNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "current password is incorrect")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	h.db.Model(&user).Update("password_hash", string(hash))

	return c.JSON(fiber.Map{"message": "password changed successfully"})
}

// ─── HELPERS ─────────────────────────────────────────────────

func generateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func mustParseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
