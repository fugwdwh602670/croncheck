package notifier_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/croncheck/internal/notifier"
)

func TestSend_Success(t *testing.T) {
	var received notifier.Alert

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notifier.New(server.URL)
	alert := notifier.Alert{
		JobName:   "backup",
		AlertType: notifier.AlertMissed,
		Message:   "job did not run within expected window",
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	if err := n.Send(alert); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.JobName != alert.JobName {
		t.Errorf("job name: got %q, want %q", received.JobName, alert.JobName)
	}
	if received.AlertType != alert.AlertType {
		t.Errorf("alert type: got %q, want %q", received.AlertType, alert.AlertType)
	}
}

func TestSend_NonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := notifier.New(server.URL)
	err := n.Send(notifier.Alert{JobName: "test", AlertType: notifier.AlertFailed})
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestSend_DefaultTimestamp(t *testing.T) {
	before := time.Now().UTC()
	var received notifier.Alert

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	n := notifier.New(server.URL)
	_ = n.Send(notifier.Alert{JobName: "nightly", AlertType: notifier.AlertMissed})

	if received.Timestamp.Before(before) {
		t.Errorf("expected timestamp >= %v, got %v", before, received.Timestamp)
	}
}
