package sla

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeHandler(t *testing.T) (*Store, http.Handler) {
	t.Helper()
	s := New()
	return s, HTTPHandler(s)
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, h := makeHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sla", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_UnknownJob(t *testing.T) {
	_, h := makeHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sla?job=ghost", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHTTPHandler_KnownJob(t *testing.T) {
	s, h := makeHandler(t)
	now := time.Now()
	s.RecordCheck("backup", now, true)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sla?job=backup", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var entry Entry
	if err := json.NewDecoder(rec.Body).Decode(&entry); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if entry.Total != 1 {
		t.Errorf("expected Total=1, got %d", entry.Total)
	}
	if entry.Hits != 1 {
		t.Errorf("expected Hits=1, got %d", entry.Hits)
	}
}

func TestHTTPHandler_AllJobs(t *testing.T) {
	s, h := makeHandler(t)
	now := time.Now()
	s.RecordCheck("job-a", now, true)
	s.RecordCheck("job-b", now, false)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sla", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var all map[string]Entry
	if err := json.NewDecoder(rec.Body).Decode(&all); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(all))
	}
}

func TestHTTPHandler_ContentType(t *testing.T) {
	_, h := makeHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sla", nil)
	h.ServeHTTP(rec, req)
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}
