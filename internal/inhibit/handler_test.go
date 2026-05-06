package inhibit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeHandler() (*Store, http.Handler) {
	s := makeStore()
	return s, HTTPHandler(s)
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/inhibit", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_EmptyUnhealthy(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/inhibit", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp statusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(resp.Rules))
	}
	if len(resp.Unhealthy) != 0 {
		t.Errorf("expected no unhealthy sources, got %d", len(resp.Unhealthy))
	}
}

func TestHTTPHandler_WithUnhealthySource(t *testing.T) {
	s, h := makeHandler()
	s.SetUnhealthy("db-backup")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/inhibit", nil))

	var resp statusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, ok := resp.Unhealthy["db-backup"]; !ok {
		t.Error("expected db-backup in unhealthy sources")
	}
}

func TestHTTPHandler_ContentType(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/inhibit", nil))
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}
