package trend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeHandler() (*Store, http.HandlerFunc) {
	s := New(10 * time.Minute)
	s.now = func() time.Time { return fixedNow }
	return s, HTTPHandler(s)
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPost, "/trend", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_UnknownJob(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/trend?job=ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHTTPHandler_KnownJob(t *testing.T) {
	s, h := makeHandler()
	_ = s.Record("sync", false)
	_ = s.Record("sync", true)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/trend?job=sync", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var e Entry
	if err := json.NewDecoder(rec.Body).Decode(&e); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if e.Job != "sync" {
		t.Errorf("expected job=sync, got %q", e.Job)
	}
	if e.Total != 2 || e.Misses != 1 {
		t.Errorf("unexpected totals: total=%d misses=%d", e.Total, e.Misses)
	}
}

func TestHTTPHandler_AllJobs(t *testing.T) {
	s, h := makeHandler()
	_ = s.Record("a", false)
	_ = s.Record("b", true)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/trend", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestHTTPHandler_ContentType(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/trend", nil))
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}
