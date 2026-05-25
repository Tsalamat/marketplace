package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"student-marketplace/internal/config"
	"student-marketplace/internal/models"
	jwtpkg "student-marketplace/pkg/jwt"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

const (
	OAuthStateCookie   = "oauth_state"
	OAuthPKCECookie    = "oauth_pkce"
	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
	CookieMaxAge       = 900 // 15 minutes for state, 30 days for tokens
	TokenCookieMaxAge  = 86400 * 30
)

type OAuthHandler struct {
	db     *gorm.DB
	google *oauth2.Config
}

func NewOAuthHandler(db *gorm.DB) *OAuthHandler {
	cfg := config.Cfg.Google
	return &OAuthHandler{
		db: db,
		google: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.CallbackURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

// generateCodeChallenge creates a PKCE code challenge from a code verifier
func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// GET /api/v1/auth/google  →  redirect to Google consent screen
// Pass ?mobile=true when calling from the mobile app.
func (h *OAuthHandler) Redirect(c *fiber.Ctx) error {
	if h.google.ClientID == "" {
		return fiber.NewError(fiber.StatusNotImplemented, "Google OAuth is not configured")
	}

	state := generateToken(32)
	codeVerifier := generateToken(48)
	codeChallenge := generateCodeChallenge(codeVerifier)

	secure := config.Cfg.App.Env == "production"

	c.Cookie(&fiber.Cookie{
		Name:     OAuthStateCookie,
		Value:    state,
		MaxAge:   CookieMaxAge,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
	})
	c.Cookie(&fiber.Cookie{
		Name:     OAuthPKCECookie,
		Value:    codeVerifier,
		MaxAge:   CookieMaxAge,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
	})

	// Track whether the request came from the mobile app so the callback
	// can redirect to the app's custom URL scheme instead of the web frontend.
	if c.Query("mobile") == "true" {
		c.Cookie(&fiber.Cookie{
			Name:     "oauth_mobile",
			Value:    "true",
			MaxAge:   CookieMaxAge,
			HTTPOnly: true,
			Secure:   secure,
			SameSite: "Lax",
			Path:     "/",
		})
	}

	authURL := h.google.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	return c.Redirect(authURL, fiber.StatusTemporaryRedirect)
}

// GET /api/v1/auth/google/callback  →  exchange code, create session, redirect to frontend
func (h *OAuthHandler) Callback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing authorization code")
	}

	// Verify CSRF state
	cookieState := c.Cookies(OAuthStateCookie)
	queryState := c.Query("state")
	if cookieState == "" || queryState == "" || cookieState != queryState {
		return fiber.NewError(fiber.StatusBadRequest, "invalid CSRF state - please try again")
	}

	// Get PKCE code verifier from cookie
	codeVerifier := c.Cookies(OAuthPKCECookie)
	if codeVerifier == "" {
		return fiber.NewError(fiber.StatusBadRequest, "PKCE verifier not found")
	}

	// Clear the state and PKCE cookies
	c.Cookie(&fiber.Cookie{
		Name:     OAuthStateCookie,
		Value:    "",
		MaxAge:   -1,
		HTTPOnly: true,
		Path:     "/",
	})
	c.Cookie(&fiber.Cookie{
		Name:     OAuthPKCECookie,
		Value:    "",
		MaxAge:   -1,
		HTTPOnly: true,
		Path:     "/",
	})

	// Exchange code for token with PKCE verifier
	token, err := h.google.Exchange(context.Background(), code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)

	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "failed to exchange token: "+err.Error())
	}

	client := h.google.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return fiber.ErrInternalServerError
	}
	defer resp.Body.Close()

	var info struct {
		ID         string `json:"id"`
		Email      string `json:"email"`
		Name       string `json:"name"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Picture    string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil || info.Email == "" {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get user info from Google")
	}

	// Find or create user
	var user models.User
	err = h.db.Preload("Profile").
		Where("google_id = ? OR (email = ? AND email != '')", info.ID, info.Email).
		First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// New Google user — create account
		username := googleUsername(info.Email)
		randomPass := generateToken(32)
		user = models.User{
			Email:         info.Email,
			Username:      username,
			PasswordHash:  randomPass,
			Role:          models.RoleBuyer,
			GoogleID:      &info.ID,
			EmailVerified: true,
			IsActive:      true,
		}
		if err := h.db.Create(&user).Error; err != nil {
			return fiber.ErrInternalServerError
		}
		profile := models.Profile{
			UserID:    user.ID,
			FirstName: info.GivenName,
			LastName:  info.FamilyName,
			AvatarURL: info.Picture,
		}
		h.db.Create(&profile)
		user.Profile = &profile
	} else if err != nil {
		return fiber.ErrInternalServerError
	} else {
		// Existing user — link Google ID if missing
		if user.GoogleID == nil {
			h.db.Model(&user).Update("google_id", info.ID)
		}
		// Update avatar if profile has none
		if user.Profile != nil && user.Profile.AvatarURL == "" && info.Picture != "" {
			h.db.Model(user.Profile).Update("avatar_url", info.Picture)
		}
	}

	if !user.IsActive {
		frontendURL := config.Cfg.App.FrontendURL
		return c.Redirect(frontendURL+"/login?error=account_banned", fiber.StatusTemporaryRedirect)
	}

	tokens, err := jwtpkg.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return fiber.ErrInternalServerError
	}

	secure := config.Cfg.App.Env == "production"

	// If the request originated from the mobile app, redirect to the app's
	// custom URL scheme with tokens in query params. The app intercepts this
	// URL via an intent filter and stores the tokens in secure storage.
	isMobile := c.Cookies("oauth_mobile") == "true"
	c.Cookie(&fiber.Cookie{Name: "oauth_mobile", Value: "", MaxAge: -1, HTTPOnly: true, Path: "/"})

	if isMobile {
		params := url.Values{
			"at": {tokens.AccessToken},
			"rt": {tokens.RefreshToken},
		}
		return c.Redirect("studentmarketplace:///oauth-callback?"+params.Encode(), fiber.StatusTemporaryRedirect)
	}

	// Web: store tokens in HTTPOnly cookies and redirect to the SPA callback.
	c.Cookie(&fiber.Cookie{
		Name:     AccessTokenCookie,
		Value:    tokens.AccessToken,
		MaxAge:   int(config.Cfg.JWT.AccessExpire.Seconds()),
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
	})
	c.Cookie(&fiber.Cookie{
		Name:     RefreshTokenCookie,
		Value:    tokens.RefreshToken,
		MaxAge:   TokenCookieMaxAge,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
	})

	frontendURL := config.Cfg.App.FrontendURL
	return c.Redirect(frontendURL+"/auth/callback?success=true", fiber.StatusTemporaryRedirect)
}

// POST /api/v1/auth/google/id-token
// Verifies a Google ID token obtained from the native google_sign_in SDK.
// Returns JWT tokens directly (no redirect needed — mobile-only endpoint).
func (h *OAuthHandler) VerifyIDToken(c *fiber.Ctx) error {
	var body struct {
		IDToken string `json:"id_token"`
	}
	if err := c.BodyParser(&body); err != nil || body.IDToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id_token required")
	}

	// Verify the ID token with Google's public tokeninfo endpoint.
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(body.IDToken))
	if err != nil || resp.StatusCode != 200 {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid id_token")
	}
	defer resp.Body.Close()

	var info struct {
		Sub        string `json:"sub"`
		Email      string `json:"email"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Picture    string `json:"picture"`
		Aud        string `json:"aud"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil || info.Email == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "could not decode token info")
	}

	// Find or create user (same logic as the web OAuth callback).
	var user models.User
	err = h.db.Preload("Profile").
		Where("google_id = ? OR (email = ? AND email != '')", info.Sub, info.Email).
		First(&user).Error

	if err == gorm.ErrRecordNotFound {
		username := googleUsername(info.Email)
		randomPass := generateToken(32)
		user = models.User{
			Email:         info.Email,
			Username:      username,
			PasswordHash:  randomPass,
			Role:          models.RoleBuyer,
			GoogleID:      &info.Sub,
			EmailVerified: true,
			IsActive:      true,
		}
		if err := h.db.Create(&user).Error; err != nil {
			return fiber.ErrInternalServerError
		}
		profile := models.Profile{
			UserID:    user.ID,
			FirstName: info.GivenName,
			LastName:  info.FamilyName,
			AvatarURL: info.Picture,
		}
		h.db.Create(&profile)
		user.Profile = &profile
	} else if err != nil {
		return fiber.ErrInternalServerError
	} else {
		if user.GoogleID == nil {
			h.db.Model(&user).Update("google_id", info.Sub)
		}
		if user.Profile != nil && user.Profile.AvatarURL == "" && info.Picture != "" {
			h.db.Model(user.Profile).Update("avatar_url", info.Picture)
		}
	}

	if !user.IsActive {
		return fiber.NewError(fiber.StatusForbidden, "account is banned")
	}

	tokens, err := jwtpkg.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(AuthResponse{User: &user, TokenPair: tokens})
}

func googleUsername(email string) string {
	base := strings.Split(email, "@")[0]
	base = strings.NewReplacer(".", "_", "-", "_", "+", "_").Replace(base)
	if len(base) > 20 {
		base = base[:20]
	}
	// Make unique with short random suffix
	return base + "_" + generateToken(3)
}

// GET /api/v1/auth/oauth-tokens - Retrieve tokens from secure cookies (for SPA)
func (h *OAuthHandler) GetTokens(c *fiber.Ctx) error {
	accessToken := c.Cookies(AccessTokenCookie)
	refreshToken := c.Cookies(RefreshTokenCookie)

	if accessToken == "" {
		return fiber.ErrUnauthorized
	}

	// Calculate access token expiry
	expiresIn := int(config.Cfg.JWT.AccessExpire.Seconds())

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    expiresIn,
		"token_type":    "Bearer",
	})
}
