package repository

import (
	"database/sql"
	"fmt"
	"time"

	"hyttekos/internal/models"
)

type TemperatureRepository struct {
	db *sql.DB
}

func NewTemperatureRepository(db *sql.DB) *TemperatureRepository {
	return &TemperatureRepository{db: db}
}

// Store saves a new temperature reading
func (r *TemperatureRepository) Store(temperature float64, timestamp time.Time) error {
	query := `
		INSERT INTO temperature_readings (temperature, timestamp)
		VALUES ($1, $2)
	`
	_, err := r.db.Exec(query, temperature, timestamp)
	if err != nil {
		return fmt.Errorf("failed to store temperature reading: %w", err)
	}
	return nil
}

// GetLatest returns the most recent temperature reading
func (r *TemperatureRepository) GetLatest() (*models.TemperatureReading, error) {
	query := `
		SELECT id, temperature, timestamp, created_at
		FROM temperature_readings
		ORDER BY timestamp DESC
		LIMIT 1
	`
	
	var reading models.TemperatureReading
	err := r.db.QueryRow(query).Scan(
		&reading.ID,
		&reading.Temperature,
		&reading.Timestamp,
		&reading.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No readings available
		}
		return nil, fmt.Errorf("failed to get latest temperature reading: %w", err)
	}
	
	return &reading, nil
}

// GetHistory returns temperature readings within a time range
func (r *TemperatureRepository) GetHistory(from, to time.Time, limit int) ([]models.TemperatureReading, error) {
	query := `
		SELECT id, temperature, timestamp, created_at
		FROM temperature_readings
		WHERE timestamp >= $1 AND timestamp <= $2
		ORDER BY timestamp DESC
		LIMIT $3
	`
	
	rows, err := r.db.Query(query, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get temperature history: %w", err)
	}
	defer rows.Close()
	
	var readings []models.TemperatureReading
	for rows.Next() {
		var reading models.TemperatureReading
		err := rows.Scan(
			&reading.ID,
			&reading.Temperature,
			&reading.Timestamp,
			&reading.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan temperature reading: %w", err)
		}
		readings = append(readings, reading)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating temperature readings: %w", err)
	}
	
	return readings, nil
}

// Cleanup removes temperature readings older than the specified duration
func (r *TemperatureRepository) Cleanup(olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM temperature_readings
		WHERE timestamp < $1
	`
	
	cutoff := time.Now().Add(-olderThan)
	result, err := r.db.Exec(query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old temperature readings: %w", err)
	}
	
	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get cleanup count: %w", err)
	}
	
	return count, nil
}