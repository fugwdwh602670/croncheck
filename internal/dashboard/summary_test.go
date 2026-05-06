package dashboard_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/croncheck/internal/dashboard"
	"github.com/example/croncheck/internal/store"
)

type stubStore struct {
	states []store.JobState
}

func (s *stubStore) All() []store.JobState { return s.states }

func makeHandler(states []store.JobState) http.HandlerFunc {
	return dashboard.HTTPHandler(&stubStore{states: states})
}

func TestHTTPHandler_EmptyStore(t *testing.T) {
	h := makeHandler(nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dashboard", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var s dashboard.Summary
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if s.Total != 0 {
		t.Errorf("expected Total=0, got %d", s.Total)
	}
}

func TestHTTPHandler_CountsHealthy(t *testing.T) {
	now := time.Now()
	states := []store.JobState{
		{Name: "job-a", Healthy: true, LastSeen: now},
		{Name: "job-b", Healthy: false, LastSeen: now, MissedCount: 2},
		{Name: "job-c", Healthy: false, LastSeen: time.Time{}},
	}
	h := makeHandler(states)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dashboard", nil))

	var s dashboard.Summary
	if err := json.NewDecoder(rec.Body).Decode(&s); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if s.Total != 3 {
		t.Errorf("Total: want 3, got %d", s.Total)
	}
	if s.Healthy != 1 {
		t.Errorf("Healthy: want 1, got %d", s.Healthy)
	}
	if s.Unhealthy != 1 {
		t.Errorf("Unhealthy: want 1, got %d", s.Unhealthy)
	}
	if s.NeverSeen != 1 {
		t.Errorf("NeverSeen: want 1, got %d", s.NeverSeen)
	}
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	h := makeHandler(nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/dashboard", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_ContentType(t *testing.T) {
	h := makeHandler(nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dashboard", nil))
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}
