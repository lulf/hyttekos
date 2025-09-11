package handlers

import (
	"net/http"
	"time"

	"hyttekos/internal/models"
	"hyttekos/internal/repository"

	"github.com/gin-gonic/gin"
)

type TemperatureHandler struct {
	tempRepo *repository.TemperatureRepository
}

func NewTemperatureHandler(tempRepo *repository.TemperatureRepository) *TemperatureHandler {
	return &TemperatureHandler{
		tempRepo: tempRepo,
	}
}

// SubmitTemperature handles temperature submissions from sensors
func (h *TemperatureHandler) SubmitTemperature(c *gin.Context) {
	var submission models.TemperatureSubmission
	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, models.APIError{
			Error:   "validation_error",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Use provided timestamp or current time
	timestamp := submission.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	// Validate temperature range (reasonable for cabin temperatures)
	if submission.Value < -50 || submission.Value > 60 {
		c.JSON(http.StatusBadRequest, models.APIError{
			Error:   "invalid_temperature",
			Message: "Temperature must be between -50 and 60 degrees Celsius",
		})
		return
	}

	// Store temperature reading
	if err := h.tempRepo.Store(submission.Value, timestamp); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIError{
			Error:   "database_error",
			Message: "Failed to store temperature reading",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Temperature recorded successfully",
	})
}

