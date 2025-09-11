package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"hyttekos/internal/akiles"
	"hyttekos/internal/middleware"
	"hyttekos/internal/models"
	"hyttekos/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

type AuthHandler struct {
	akilesClient  *akiles.Client
	sessionStore  sessions.Store
	sessionRepo   *repository.SessionRepository
	redirectURL   string
}

func NewAuthHandler(client *akiles.Client, store sessions.Store, sessionRepo *repository.SessionRepository, redirectURL string) *AuthHandler {
	return &AuthHandler{
		akilesClient: client,
		sessionStore: store,
		sessionRepo:  sessionRepo,
		redirectURL:  redirectURL,
	}
}

// Login initiates the OAuth2 flow
func (h *AuthHandler) Login(c *gin.Context) {
	// Generate a random state for CSRF protection
	state := generateRandomState()
	
	// Store state in session
	session, err := h.sessionStore.Get(c.Request, middleware.SessionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "session_error",
			"message": "Failed to create session",
		})
		return
	}
	
	session.Values["oauth_state"] = state
	if err := session.Save(c.Request, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "session_error",
			"message": "Failed to save session",
		})
		return
	}

	// Redirect to OAuth provider
	authURL := h.akilesClient.GetAuthURL(state, h.redirectURL)
	c.Redirect(http.StatusFound, authURL)
}

// Callback handles the OAuth2 callback
func (h *AuthHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_code",
			"message": "Authorization code is required",
		})
		return
	}

	// Validate state parameter
	session, err := h.sessionStore.Get(c.Request, middleware.SessionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "session_error",
			"message": "Failed to get session",
		})
		return
	}
	
	expectedState, ok := session.Values["oauth_state"].(string)
	if !ok || expectedState != state {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_state",
			"message": "Invalid state parameter",
		})
		return
	}

	// Exchange code for tokens
	token, err := h.akilesClient.ExchangeCode(c.Request.Context(), code, h.redirectURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "token_exchange_failed",
			"message": fmt.Sprintf("Failed to exchange code for token: %v", err),
		})
		return
	}

	// Generate session ID
	sessionID := generateRandomState()
	
	// Store session in database
	dbSession := &models.Session{
		ID:           sessionID,
		UserID:       "user", // In a real app, you'd get user info from Akiles
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
	}
	
	if err := h.sessionRepo.Store(dbSession); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "session_store_failed",
			"message": fmt.Sprintf("Failed to store session: %v", err),
		})
		return
	}

	// Store session ID in cookie
	session.Values["session_id"] = sessionID
	delete(session.Values, "oauth_state") // Clean up state
	
	if err := session.Save(c.Request, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "session_save_failed",
			"message": fmt.Sprintf("Failed to save session cookie: %v", err),
		})
		return
	}

	// Redirect to main application
	c.Redirect(http.StatusFound, "/")
}

// Logout clears the session
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, exists := c.Get("session_id")
	if exists {
		if id, ok := sessionID.(string); ok {
			// Remove from database
			h.sessionRepo.Delete(id)
		}
	}

	// Clear session cookie
	session, err := h.sessionStore.Get(c.Request, middleware.SessionName)
	if err == nil {
		delete(session.Values, "session_id")
		session.Save(c.Request, c.Writer)
	}

	// Check if request is HTMX
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", "/login")
		c.Status(http.StatusOK)
		return
	}

	// Check if request expects JSON
	if c.GetHeader("Accept") == "application/json" {
		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
		return
	}

	// Regular redirect
	c.Redirect(http.StatusFound, "/login")
}

func generateRandomState() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}