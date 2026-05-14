package snapshot

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeHandler() (*Store, http.Handler) {
	s := New(10)
	s.now = func() time.Time { return time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC) }
	src := &mockSource{states: []JobState{
		{Job: "daily-report", Healthy: true, MissedCount: 0},
	}}
	return s, HTTPHandler(s, src)
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/snapshots", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_ListEmpty(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/snapshots", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result []Snapshot
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty list, got %d", len(result))
	}
}

func TestHTTPHandler_CaptureAndList(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/snapshots?id=test-snap", nil))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/snapshots", nil))
	var result []Snapshot
	json.NewDecoder(rec2.Body).Decode(&result)
	if len(result) != 1 || result[0].ID != "test-snap" {
		t.Fatalf("unexpected snapshots: %+v", result)
	}
}

func TestHTTPHandler_GetByID(t *testing.T) {
	_, h := makeHandler()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/snapshots?id=snap-x", nil))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/snapshots/snap-x", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var snap Snapshot
	json.NewDecoder(rec.Body).Decode(&snap)
	if snap.ID != "snap-x" {
		t.Fatalf("expected snap-x, got %s", snap.ID)
	}
}

func TestHTTPHandler_GetUnknownID(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/snapshots/ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHTTPHandler_PostMissingID(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/snapshots", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
