package notifier

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
	AlertMissed AlertType = "missed"
	AlertFailed AlertType = "failed"
)

// Alert holds the data for a single notification.
type Alert struct {
	JobName   string    `json:"job_name"`
	AlertType AlertType `json:"alert_type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Notifier sends alerts to a configured webhook URL.
type Notifier struct {
	webhookURL string
	client     *http.Client
}

// New creates a new Notifier with the given webhook URL.
func New(webhookURL string) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send dispatches an alert to the configured webhook.
func (n *Notifier) Send(alert Alert) error {
	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now().UTC()
	}

	payload, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("notifier: marshal alert: %w", err)
	}

	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("notifier: post alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notifier: unexpected status %d from webhook", resp.StatusCode)
	}

	return nil
}
