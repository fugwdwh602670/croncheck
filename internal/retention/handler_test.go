package retention_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/croncheck/internal/retention"
)

func TestHTTPHandler_ReturnsConfig(t *testing.T) {
	sr := retention.StatusReporter{
		TTL:      2 * time.Hour,
		Interval: 15 * time.Minute,
	}
	h := retention.HTTPHandler(sr)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/retention", nil)
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["ttl"] != "2h0m0s" {
		t.Errorf("unexpected ttl: %s", body["ttl"])
	}
	if body["interval"] != "15m0s" {
		t.Errorf("unexpected interval: %s", body["interval"])
	}
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	sr := retention.StatusReporter{TTL: time.Hour, Interval: time.Minute}
	h := retention.HTTPHandler(sr)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/retention", nil)
	h(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_ContentType(t *testing.T) {
	sr := retention.StatusReporter{TTL: time.Hour, Interval: time.Minute}
	h := retention.HTTPHandler(sr)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/retention", nil)
	h(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}
