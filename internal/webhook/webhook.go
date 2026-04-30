// Package webhook provides functionality for sending alert notifications
// to configured webhook endpoints when cron job anomalies are detected.
package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AlertType represents the kind of alert being sent.
type AlertType string

const (
	AlertMissed     AlertType = "missed"
	AlertLongRunning AlertType = "long_running"
)

// Payload is the JSON body sent to the webhook endpoint.
type Payload struct {
	JobName   string    `json:"job_name"`
	AlertType AlertType `json:"alert_type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Client sends webhook notifications.
type Client struct {
	URL        string
	HTTPClient *http.Client
}

// NewClient creates a new webhook Client with a default timeout.
func NewClient(url string) *Client {
	return &Client{
		URL: url,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send serializes the payload and posts it to the webhook URL.
func (c *Client) Send(p Payload) error {
	if p.Timestamp.IsZero() {
		p.Timestamp = time.Now().UTC()
	}

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	resp, err := c.HTTPClient.Post(c.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status code %d", resp.StatusCode)
	}

	return nil
}
