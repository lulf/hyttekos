package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"hyttekos/internal/akiles"
	"hyttekos/internal/models"
	"hyttekos/internal/repository"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type APIHandler struct {
	akilesClient *akiles.Client
	tempRepo     *repository.TemperatureRepository
	sessionRepo  *repository.SessionRepository
}

func NewAPIHandler(client *akiles.Client, tempRepo *repository.TemperatureRepository, sessionRepo *repository.SessionRepository) *APIHandler {
	return &APIHandler{
		akilesClient: client,
		tempRepo:     tempRepo,
		sessionRepo:  sessionRepo,
	}
}

// GetStatus returns the current status of all systems
func (h *APIHandler) GetStatus(c *gin.Context) {
	accessToken, exists := c.Get("access_token")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIError{
			Error:   "missing_token",
			Message: "Access token not found",
		})
		return
	}

	token := &oauth2.Token{
		AccessToken: accessToken.(string),
		TokenType:   "Bearer",
	}

	// Get heating status
	heatingState, err := h.akilesClient.GetHeatingState(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Error:   "akiles_api_error",
			Message: fmt.Sprintf("Failed to get heating status: %v", err),
		})
		return
	}

	// Get hot water status
	hotWaterState, err := h.akilesClient.GetHotWaterState(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Error:   "akiles_api_error",
			Message: fmt.Sprintf("Failed to get hot water status: %v", err),
		})
		return
	}

	// Get latest temperature
	latestTemp, err := h.tempRepo.GetLatest()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Error:   "database_error",
			Message: fmt.Sprintf("Failed to get temperature: %v", err),
		})
		return
	}

	status := models.SystemStatus{
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
		// Return empty temperature reading if none available
		status.Temperature = models.TemperatureReading{
			Temperature: 0,
			Timestamp:   time.Now(),
		}
	}

	c.JSON(http.StatusOK, status)
}

// SetHeating controls the heating system
func (h *APIHandler) SetHeating(c *gin.Context) {
	var action models.GadgetAction
	if err := c.ShouldBindJSON(&action); err != nil {
		c.JSON(http.StatusBadRequest, models.APIError{
			Error:   "validation_error",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	accessToken, _ := c.Get("access_token")
	token := &oauth2.Token{
		AccessToken: accessToken.(string),
		TokenType:   "Bearer",
	}

	enabled := action.Action == "on"
	_, err := h.akilesClient.SetHeating(c.Request.Context(), token, enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Error:   "akiles_api_error",
			Message: fmt.Sprintf("Failed to control heating: %v", err),
		})
		return
	}

	// Get updated status
	heatingState, err := h.akilesClient.GetHeatingState(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Error:   "akiles_api_error",
			Message: fmt.Sprintf("Failed to get updated heating status: %v", err),
		})
		return
	}

	response := models.GadgetActionResponse{
		Success: true,
		Heating: &models.GadgetStatus{
			Enabled:     heatingState.State == "open",
			LastUpdated: heatingState.UpdatedAt,
		},
	}

	c.JSON(http.StatusOK, response)
}

// SetHotWater controls the hot water system
func (h *APIHandler) SetHotWater(c *gin.Context) {
	var action models.GadgetAction
	if err := c.ShouldBindJSON(&action); err != nil {
		c.JSON(http.StatusBadRequest, models.APIError{
			Error:   "validation_error",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	accessToken, _ := c.Get("access_token")
	token := &oauth2.Token{
		AccessToken: accessToken.(string),
		TokenType:   "Bearer",
	}

	enabled := action.Action == "on"
	_, err := h.akilesClient.SetHotWater(c.Request.Context(), token, enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Error:   "akiles_api_error",
			Message: fmt.Sprintf("Failed to control hot water: %v", err),
		})
		return
	}

	// Get updated status
	hotWaterState, err := h.akilesClient.GetHotWaterState(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Error:   "akiles_api_error",
			Message: fmt.Sprintf("Failed to get updated hot water status: %v", err),
		})
		return
	}

	response := models.GadgetActionResponse{
		Success: true,
		HotWater: &models.GadgetStatus{
			Enabled:     hotWaterState.State == "open",
			LastUpdated: hotWaterState.UpdatedAt,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetTemperature returns the current temperature
func (h *APIHandler) GetTemperature(c *gin.Context) {
	latestTemp, err := h.tempRepo.GetLatest()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Error:   "database_error",
			Message: fmt.Sprintf("Failed to get temperature: %v", err),
		})
		return
	}

	if latestTemp == nil {
		c.JSON(http.StatusNotFound, models.APIError{
			Error:   "no_data",
			Message: "No temperature readings available",
		})
		return
	}

	c.JSON(http.StatusOK, latestTemp)
}

// GetTemperatureHistory returns temperature history
func (h *APIHandler) GetTemperatureHistory(c *gin.Context) {
	// Parse query parameters
	fromStr := c.DefaultQuery("from", "")
	toStr := c.DefaultQuery("to", "")
	limitStr := c.DefaultQuery("limit", "100")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 1000 {
		limit = 100
	}

	// Default to last 24 hours if no time range specified
	to := time.Now()
	from := to.Add(-24 * time.Hour)

	if fromStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = parsed
		}
	}

	if toStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = parsed
		}
	}

	readings, err := h.tempRepo.GetHistory(from, to, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Error:   "database_error",
			Message: fmt.Sprintf("Failed to get temperature history: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"readings": readings})
}