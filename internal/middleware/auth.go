package middleware

import (
	"fmt"
	"net/http"
	"time"

	"hyttekos/internal/akiles"
	"hyttekos/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

const SessionName = "hyttekos-session"

// RequireAuth middleware checks for valid authentication
func RequireAuth(store sessions.Store, sessionRepo *repository.SessionRepository, akilesClient ...*akiles.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := store.Get(c.Request, SessionName)
		if err != nil {
			redirectToLogin(c)
			return
		}

		sessionID, ok := session.Values["session_id"].(string)
		if !ok || sessionID == "" {
			redirectToLogin(c)
			return
		}

		// Validate session in database
		dbSession, err := sessionRepo.GetByID(sessionID)
		if err != nil || dbSession == nil {
			// Clear invalid session
			delete(session.Values, "session_id")
			session.Save(c.Request, c.Writer)
			redirectToLogin(c)
			return
		}

		// If token is expired, try to refresh it
		if dbSession.ExpiresAt.Before(time.Now()) && dbSession.RefreshToken != "" && len(akilesClient) > 0 && akilesClient[0] != nil {
			oldToken := &oauth2.Token{
				AccessToken:  dbSession.AccessToken,
				RefreshToken: dbSession.RefreshToken,
				Expiry:       dbSession.ExpiresAt,
			}
			newToken, err := akilesClient[0].RefreshToken(c.Request.Context(), oldToken)
			if err != nil {
				fmt.Printf("RequireAuth: Failed to refresh token: %v\n", err)
				delete(session.Values, "session_id")
				session.Save(c.Request, c.Writer)
				sessionRepo.Delete(sessionID)
				redirectToLogin(c)
				return
			}
			// Update session in database with new tokens
			if err := sessionRepo.UpdateTokens(sessionID, newToken.AccessToken, newToken.RefreshToken, newToken.Expiry); err != nil {
				fmt.Printf("RequireAuth: Failed to update refreshed tokens: %v\n", err)
			}
			dbSession.AccessToken = newToken.AccessToken
			dbSession.RefreshToken = newToken.RefreshToken
			dbSession.ExpiresAt = newToken.Expiry
		}

		// Store session info in context for handlers
		c.Set("session_id", sessionID)
		c.Set("user_id", dbSession.UserID)
		c.Set("access_token", dbSession.AccessToken)
		c.Set("refresh_token", dbSession.RefreshToken)

		c.Next()
	}
}

func redirectToLogin(c *gin.Context) {
	// Check if request is HTMX
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", "/login")
		c.Status(http.StatusUnauthorized)
		c.Abort()
		return
	}

	// Check if request expects JSON
	if c.GetHeader("Accept") == "application/json" || c.GetHeader("Content-Type") == "application/json" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Valid authentication token required",
		})
		c.Abort()
		return
	}

	// Regular redirect for browser
	c.Redirect(http.StatusFound, "/login")
	c.Abort()
}