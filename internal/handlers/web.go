package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"hyttekos/internal/akiles"
	"hyttekos/internal/models"
	"hyttekos/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type WebHandler struct {
	akilesClient *akiles.Client
	tempRepo     *repository.TemperatureRepository
	sessionRepo  *repository.SessionRepository
	sessionStore sessions.Store
}

func NewWebHandler(client *akiles.Client, tempRepo *repository.TemperatureRepository, sessionRepo *repository.SessionRepository, sessionStore sessions.Store) *WebHandler {
	return &WebHandler{
		akilesClient: client,
		tempRepo:     tempRepo,
		sessionRepo:  sessionRepo,
		sessionStore: sessionStore,
	}
}

// Index serves the main dashboard page with all data
func (h *WebHandler) Index(c *gin.Context) {
	accessToken, exists := c.Get("access_token")
	if !exists {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Autentiseringstoken ikke funnet",
		})
		return
	}

	token := &oauth2.Token{
		AccessToken: accessToken.(string),
		TokenType:   "Bearer",
	}

	// Get system status
	status, err := h.getSystemStatus(c, token)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to get system status: %v", err),
		})
		return
	}

	// Format last updated time
	lastUpdated := time.Now().Format("15:04")

	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":       "Hyttekos - Hyttestyring",
		"heating":     status.Heating,
		"hotWater":    status.HotWater,
		"temperature": status.Temperature,
		"lastUpdated": lastUpdated,
	})
}

// LoginPage serves the login page
func (h *WebHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Innlogging - Hyttekos",
	})
}


// ToggleHeating handles heating system toggle requests
func (h *WebHandler) ToggleHeating(c *gin.Context) {
	accessToken, _ := c.Get("access_token")
	token := &oauth2.Token{
		AccessToken: accessToken.(string),
		TokenType:   "Bearer",
	}

	// Get current state
	heatingState, err := h.akilesClient.GetHeatingState(c.Request.Context(), token)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to get heating status: %v", err),
		})
		return
	}

	// Toggle the state
	newEnabled := heatingState.StateID != "closed"
	_, err = h.akilesClient.SetHeating(c.Request.Context(), token, newEnabled)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to control heating: %v", err),
		})
		return
	}

	// Redirect back to main page
	c.Redirect(http.StatusSeeOther, "/")
}

// ToggleHotWater handles hot water system toggle requests
func (h *WebHandler) ToggleHotWater(c *gin.Context) {
	accessToken, _ := c.Get("access_token")
	token := &oauth2.Token{
		AccessToken: accessToken.(string),
		TokenType:   "Bearer",
	}

	// Get current state
	hotWaterState, err := h.akilesClient.GetHotWaterState(c.Request.Context(), token)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to get hot water status: %v", err),
		})
		return
	}

	// Toggle the state
	newEnabled := hotWaterState.StateID != "closed"
	_, err = h.akilesClient.SetHotWater(c.Request.Context(), token, newEnabled)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to control hot water: %v", err),
		})
		return
	}

	// Redirect back to main page
	c.Redirect(http.StatusSeeOther, "/")
}


func (h *WebHandler) getSystemStatus(c *gin.Context, token *oauth2.Token) (*models.SystemStatus, error) {
	// Get heating status
	log.Printf("getSystemStatus: Getting heating state")
	heatingState, err := h.akilesClient.GetHeatingState(c.Request.Context(), token)
	if err != nil {
		log.Printf("getSystemStatus: Error getting heating state: %v", err)
		return nil, fmt.Errorf("failed to get heating status: %w", err)
	}
	log.Printf("getSystemStatus: Heating state retrieved: %+v", heatingState)

	// Get hot water status
	log.Printf("getSystemStatus: Getting hot water state")
	hotWaterState, err := h.akilesClient.GetHotWaterState(c.Request.Context(), token)
	if err != nil {
		log.Printf("getSystemStatus: Error getting hot water state: %v", err)
		return nil, fmt.Errorf("failed to get hot water status: %w", err)
	}
	log.Printf("getSystemStatus: Hot water state retrieved: %+v", hotWaterState)

	// Get latest temperature
	log.Printf("getSystemStatus: Getting latest temperature")
	latestTemp, err := h.tempRepo.GetLatest()
	if err != nil {
		log.Printf("getSystemStatus: Error getting temperature: %v", err)
		return nil, fmt.Errorf("failed to get temperature: %w", err)
	}
	log.Printf("getSystemStatus: Temperature retrieved successfully")

	status := &models.SystemStatus{
		Heating: models.GadgetStatus{
			Enabled: heatingState.StateID != "closed",
		},
		HotWater: models.GadgetStatus{
			Enabled: hotWaterState.StateID != "closed",
		},
	}

	if latestTemp != nil {
		status.Temperature = *latestTemp
	} else {
		status.Temperature = models.TemperatureReading{
			Temperature: 0,
			Timestamp:   time.Now(),
		}
	}

	log.Printf("getSystemStatus: System status created successfully")
	return status, nil
}
