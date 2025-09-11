package akiles

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	BaseURL    = "https://api.akiles.app/v2"
	AuthURL    = "https://auth.akiles.app/oauth2/auth"
	TokenURL   = "https://auth.akiles.app/oauth2/token"
	APITimeout = 30 * time.Second
)

// Client represents an Akiles API client
type Client struct {
	httpClient        *http.Client
	config            *oauth2.Config
	clientCredentials *clientcredentials.Config
	heatingGadgetID   string
	hotWaterGadgetID  string
}

// NewClient creates a new Akiles API client
func NewClient(clientID, clientSecret, heatingGadgetID, hotWaterGadgetID string) *Client {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  AuthURL,
			TokenURL: TokenURL,
		},
		Scopes: []string{"full_read_write", "offline"},
	}

	clientCredConfig := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     TokenURL,
		Scopes:       []string{"full_read_write", "offline"},
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: APITimeout,
		},
		config:            config,
		clientCredentials: clientCredConfig,
		heatingGadgetID:   heatingGadgetID,
		hotWaterGadgetID:  hotWaterGadgetID,
	}
}

// GadgetState represents the state of a gadget
type GadgetState struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	StateID   string    `json:"state_id"`
	State     string    `json:"state"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GadgetActionResponse represents the response from a gadget action
type GadgetActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	StateID string `json:"state_id"`
}

// GetAuthURL returns the OAuth2 authorization URL
func (c *Client) GetAuthURL(state string, redirectURL string) string {
	c.config.RedirectURL = redirectURL
	return c.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges an authorization code for tokens
func (c *Client) ExchangeCode(ctx context.Context, code, redirectURL string) (*oauth2.Token, error) {
	c.config.RedirectURL = redirectURL
	token, err := c.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	return token, nil
}

// GetGadgetState retrieves the current state of a gadget
func (c *Client) GetGadgetState(ctx context.Context, token *oauth2.Token, gadgetID string) (*GadgetState, error) {
	client := c.config.Client(ctx, token)

	url := fmt.Sprintf("%s/gadgets/%s", BaseURL, gadgetID)
	log.Printf("GetGadgetState: Making request to URL: %s", url)

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("GetGadgetState: HTTP request failed: %v", err)
		return nil, fmt.Errorf("failed to get gadget state: %w", err)
	}
	defer resp.Body.Close()
	log.Printf("GetGadgetState: Response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("GetGadgetState: API error response: %s", string(body))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var gadget GadgetState
	if err := json.NewDecoder(resp.Body).Decode(&gadget); err != nil {
		log.Printf("GetGadgetState: JSON decode failed: %v", err)
		return nil, fmt.Errorf("failed to decode gadget state: %w", err)
	}

	log.Printf("GetGadgetState: Successfully retrieved gadget state: %+v", gadget)
	return &gadget, nil
}

// PerformGadgetAction performs an action on a gadget (on/off)
func (c *Client) PerformGadgetAction(ctx context.Context, token *oauth2.Token, gadgetID, action string) (*GadgetActionResponse, error) {
	client := c.config.Client(ctx, token)

	url := fmt.Sprintf("%s/gadgets/%s/actions/%s", BaseURL, gadgetID, action)
	log.Printf("PerformGadgetAction: Making request to URL: %s", url)

	// Use empty JSON body
	jsonData := []byte("{}")

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("PerformGadgetAction: Failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("PerformGadgetAction: HTTP request failed: %v", err)
		return nil, fmt.Errorf("failed to perform gadget action: %w", err)
	}
	defer resp.Body.Close()
	log.Printf("PerformGadgetAction: Response status: %d", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("PerformGadgetAction: Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("PerformGadgetAction: API error response: %s", string(body))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var actionResp GadgetActionResponse
	if err := json.Unmarshal(body, &actionResp); err != nil {
		log.Printf("PerformGadgetAction: JSON decode failed: %v", err)
		return nil, fmt.Errorf("failed to decode action response: %w", err)
	}

	log.Printf("PerformGadgetAction: Successfully performed action: %+v", actionResp)
	return &actionResp, nil
}

// GetHeatingState gets the current heating system state
func (c *Client) GetHeatingState(ctx context.Context, token *oauth2.Token) (*GadgetState, error) {
	return c.GetGadgetState(ctx, token, c.heatingGadgetID)
}

// GetHotWaterState gets the current hot water system state
func (c *Client) GetHotWaterState(ctx context.Context, token *oauth2.Token) (*GadgetState, error) {
	return c.GetGadgetState(ctx, token, c.hotWaterGadgetID)
}

// SetHeating turns heating on or off
func (c *Client) SetHeating(ctx context.Context, token *oauth2.Token, enabled bool) (*GadgetActionResponse, error) {
	action := "close"
	if enabled {
		action = "open"
	}
	return c.PerformGadgetAction(ctx, token, c.heatingGadgetID, action)
}

// SetHotWater turns hot water on or off
func (c *Client) SetHotWater(ctx context.Context, token *oauth2.Token, enabled bool) (*GadgetActionResponse, error) {
	action := "close"
	if enabled {
		action = "open"
	}
	return c.PerformGadgetAction(ctx, token, c.hotWaterGadgetID, action)
}

// RefreshToken refreshes an OAuth2 token
func (c *Client) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	tokenSource := c.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	return newToken, nil
}
