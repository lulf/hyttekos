package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabasePath          string
	Port                  string
	SessionSecret         string
	OAuthRedirectURL      string
	AkilesClientID        string
	AkilesClientSecret    string
	AkilesHeatingGadgetID string
	AkilesHotWaterGadgetID string
	TempSensorUsername    string
	TempSensorPassword    string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabasePath:          getEnv("DATABASE_PATH", "./hyttekos.db"),
		Port:                  getEnv("PORT", "8080"),
		SessionSecret:         getEnv("SESSION_SECRET", "your-session-secret-here"),
		OAuthRedirectURL:      getEnv("OAUTH_REDIRECT_URL", "http://localhost:8080/auth/callback"),
		AkilesClientID:        getEnv("AKILES_CLIENT_ID", ""),
		AkilesClientSecret:    getEnv("AKILES_CLIENT_SECRET", ""),
		AkilesHeatingGadgetID: getEnv("AKILES_HEATING_GADGET_ID", ""),
		AkilesHotWaterGadgetID: getEnv("AKILES_HOTWATER_GADGET_ID", ""),
		TempSensorUsername:    getEnv("TEMP_SENSOR_USERNAME", ""),
		TempSensorPassword:    getEnv("TEMP_SENSOR_PASSWORD", ""),
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"DATABASE_PATH":             c.DatabasePath,
		"SESSION_SECRET":            c.SessionSecret,
		"AKILES_CLIENT_ID":          c.AkilesClientID,
		"AKILES_CLIENT_SECRET":      c.AkilesClientSecret,
		"AKILES_HEATING_GADGET_ID":  c.AkilesHeatingGadgetID,
		"AKILES_HOTWATER_GADGET_ID": c.AkilesHotWaterGadgetID,
		"TEMP_SENSOR_USERNAME":      c.TempSensorUsername,
		"TEMP_SENSOR_PASSWORD":      c.TempSensorPassword,
	}

	for name, value := range required {
		if value == "" {
			return fmt.Errorf("required environment variable %s is not set", name)
		}
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}