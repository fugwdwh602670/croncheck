package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/croncheck/internal/api"
	"github.com/croncheck/internal/store"
)

func newStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.New(nil)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	return s
}

func TestHandler_EmptyStore(t *testing.T) {
	s := newStore(t)
	h := api.Handler(s)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("unexpected Content-Type: %s", ct)
	}
	var statuses []api.JobStatus
	if err := json.NewDecoder(rec.Body).Decode(&statuses); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(statuses) != 0 {
		t.Errorf("expected empty slice, got %d items", len(statuses))
	}
}

func TestHandler_WithJobs(t *testing.T) {
	s := newStore(t)
	s.RecordHeartbeat("backup", time.Now())
	s.RecordHeartbeat("cleanup", time.Now())

	h := api.Handler(s)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var statuses []api.JobStatus
	if err := json.NewDecoder(rec.Body).Decode(&statuses); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(statuses) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(statuses))
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	s := newStore(t)
	h := api.Handler(s)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/status", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
