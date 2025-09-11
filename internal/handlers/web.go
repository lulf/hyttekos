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

// Index serves the main dashboard page
func (h *WebHandler) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Hyttekos - Hyttestyring",
	})
}

// LoginPage serves the login page
func (h *WebHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Innlogging - Hyttekos",
	})
}

// Dashboard returns the main dashboard HTML fragment
func (h *WebHandler) Dashboard(c *gin.Context) {
	log.Printf("Dashboard: Starting dashboard request")
	accessToken, exists := c.Get("access_token")
	if !exists {
		log.Printf("Dashboard: Authentication token not found")
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Autentiseringstoken ikke funnet",
		})
		return
	}
	log.Printf("Dashboard: Access token found, creating OAuth2 token")

	token := &oauth2.Token{
		AccessToken: accessToken.(string),
		TokenType:   "Bearer",
	}

	// Get system status
	log.Printf("Dashboard: Calling getSystemStatus")
	status, err := h.getSystemStatus(c, token)
	if err != nil {
		log.Printf("Dashboard: Error getting system status: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": fmt.Sprintf("Failed to get system status: %v", err),
		})
		return
	}
	log.Printf("Dashboard: System status retrieved successfully")

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"heating":     status.Heating,
		"hotWater":    status.HotWater,
		"temperature": status.Temperature,
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
		c.HTML(http.StatusInternalServerError, "error-fragment.html", gin.H{
			"error": fmt.Sprintf("Failed to get heating status: %v", err),
		})
		return
	}

	// Toggle the state
	newEnabled := heatingState.State != "open"
	_, err = h.akilesClient.SetHeating(c.Request.Context(), token, newEnabled)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error-fragment.html", gin.H{
			"error": fmt.Sprintf("Failed to control heating: %v", err),
		})
		return
	}

	// Get updated state
	updatedState, err := h.akilesClient.GetHeatingState(c.Request.Context(), token)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error-fragment.html", gin.H{
			"error": fmt.Sprintf("Failed to get updated heating status: %v", err),
		})
		return
	}

	heating := models.GadgetStatus{
		Enabled:     updatedState.State == "open",
		LastUpdated: updatedState.UpdatedAt,
	}

	c.HTML(http.StatusOK, "heating-card.html", gin.H{
		"heating": heating,
	})
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
		c.HTML(http.StatusInternalServerError, "error-fragment.html", gin.H{
			"error": fmt.Sprintf("Failed to get hot water status: %v", err),
		})
		return
	}

	// Toggle the state
	newEnabled := hotWaterState.State != "open"
	_, err = h.akilesClient.SetHotWater(c.Request.Context(), token, newEnabled)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error-fragment.html", gin.H{
			"error": fmt.Sprintf("Failed to control hot water: %v", err),
		})
		return
	}

	// Get updated state
	updatedState, err := h.akilesClient.GetHotWaterState(c.Request.Context(), token)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error-fragment.html", gin.H{
			"error": fmt.Sprintf("Failed to get updated hot water status: %v", err),
		})
		return
	}

	hotWater := models.GadgetStatus{
		Enabled:     updatedState.State == "open",
		LastUpdated: updatedState.UpdatedAt,
	}

	c.HTML(http.StatusOK, "hot-water-card.html", gin.H{
		"hotWater": hotWater,
	})
}

// TemperatureFragment returns the temperature display fragment
func (h *WebHandler) TemperatureFragment(c *gin.Context) {
	latestTemp, err := h.tempRepo.GetLatest()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error-fragment.html", gin.H{
			"error": fmt.Sprintf("Failed to get temperature: %v", err),
		})
		return
	}

	var temperature models.TemperatureReading
	if latestTemp != nil {
		temperature = *latestTemp
	} else {
		temperature = models.TemperatureReading{
			Temperature: 0,
			Timestamp:   time.Now(),
		}
	}

	c.HTML(http.StatusOK, "temperature-display.html", gin.H{
		"temperature": temperature,
	})
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
			Enabled:     heatingState.State == "open",
			LastUpdated: heatingState.UpdatedAt,
		},
		HotWater: models.GadgetStatus{
			Enabled:     hotWaterState.State == "open",
			LastUpdated: hotWaterState.UpdatedAt,
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