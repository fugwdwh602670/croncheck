package history

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeHandler(limit int) (*History, http.Handler) {
	h := New(limit)
	return h, HTTPHandler(h)
}

func TestHTTPHandler_UnknownJob(t *testing.T) {
	_, handler := makeHandler(10)
	req := httptest.NewRequest(http.MethodGet, "/history/unknown-job", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var events []Event
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected empty slice, got %d events", len(events))
	}
}

func TestHTTPHandler_KnownJob(t *testing.T) {
	h, handler := makeHandler(10)
	now := time.Now()
	h.Record("backup", now, true)
	h.Record("backup", now.Add(time.Minute), false)

	req := httptest.NewRequest(http.MethodGet, "/history/backup", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var events []Event
	if err := json.NewDecoder(rec.Body).Decode(&events); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestHTTPHandler_MissingJobName(t *testing.T) {
	_, handler := makeHandler(10)
	req := httptest.NewRequest(http.MethodGet, "/history/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, handler := makeHandler(10)
	req := httptest.NewRequest(http.MethodPost, "/history/backup", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_ContentType(t *testing.T) {
	_, handler := makeHandler(10)
	req := httptest.NewRequest(http.MethodGet, "/history/somejob", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}
