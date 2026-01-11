package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"vpsentinel-agent/models"
)

const (
	// Retry configuration
	maxRetries      = 5
	initialDelay    = 1 * time.Second
	maxDelay        = 60 * time.Second
	backoffMultiplier = 2.0

	// HTTP configuration
	requestTimeout = 30 * time.Second
)

// Client handles HTTPS communication with the backend
type Client struct {
	url        string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new transport client
func NewClient(url, apiKey string) *Client {
	// Ensure URL ends with / for path concatenation
	if url[len(url)-1] != '/' {
		url += "/"
	}

	return &Client{
		url:    url,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// CheckCommands checks for pending commands from the backend
func (c *Client) CheckCommands() ([]models.Command, error) {
	url := c.url + "api/agent/commands"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var commands []models.Command
	if err := json.NewDecoder(resp.Body).Decode(&commands); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return commands, nil
}

// SendCommandResponse sends a response to a command execution
func (c *Client) SendCommandResponse(commandID string, status, message string) error {
	response := models.CommandResponse{
		CommandID: commandID,
		Status:    status,
		Message:   message,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	url := c.url + "api/agent/commands/respond"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Send sends a payload to the backend with retry logic and exponential backoff
func (c *Client) Send(payload models.Payload) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate backoff delay
			delay := calculateBackoff(attempt)
			log.Printf("Retrying after %v (attempt %d/%d)", delay, attempt+1, maxRetries)
			time.Sleep(delay)
		}

		err := c.sendRequest(payload)
		if err == nil {
			if attempt > 0 {
				log.Printf("Successfully sent after %d attempts", attempt+1)
			}
			return nil
		}

		lastErr = err

		// Don't retry on authentication errors (invalid API key)
		if httpErr, ok := err.(*HTTPError); ok {
			if httpErr.StatusCode == 401 || httpErr.StatusCode == 403 {
				log.Printf("Authentication error (status %d), stopping retries", httpErr.StatusCode)
				return err
			}
		}

		log.Printf("Send attempt %d/%d failed: %v", attempt+1, maxRetries, err)
	}

	return fmt.Errorf("failed to send after %d attempts: %w", maxRetries, lastErr)
}

// sendRequest performs a single HTTP request
func (c *Client) sendRequest(payload models.Payload) error {
	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	url := c.url + "api/agent/ingest"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "VPSentinel-Agent/1.0")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body (for error messages)
	body, _ := io.ReadAll(resp.Body)

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(body),
		}
	}

	return nil
}

// calculateBackoff calculates the exponential backoff delay
func calculateBackoff(attempt int) time.Duration {
	delay := float64(initialDelay) * backoffMultiplier * float64(attempt)
	if delay > float64(maxDelay) {
		delay = float64(maxDelay)
	}
	return time.Duration(delay)
}

// HTTPError represents an HTTP error response
type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d %s: %s", e.StatusCode, e.Status, e.Body)
}
