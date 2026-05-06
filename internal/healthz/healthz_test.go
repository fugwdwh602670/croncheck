package healthz_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/croncheck/internal/healthz"
)

func TestHandler_ReturnsOK(t *testing.T) {
	h := healthz.Handler("1.2.3")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp healthz.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
	if resp.Version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %q", resp.Version)
	}
	if resp.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	if resp.Timestamp.Location() != time.UTC {
		t.Error("expected UTC timestamp")
	}
}

func TestHandler_ContentType(t *testing.T) {
	h := healthz.Handler("")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	h(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	h := healthz.Handler("")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)

	h(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_NoVersionOmitted(t *testing.T) {
	h := healthz.Handler("")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	h(rec, req)

	body := rec.Body.String()
	if contains := `"version"`; containsStr(body, contains) {
		t.Errorf("expected version field to be omitted when empty, body: %s", body)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s[1:], sub) || s[:len(sub)] == sub)
}
