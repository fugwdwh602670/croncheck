package runlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeHandler(t *testing.T) (*RunLog, http.Handler) {
	t.Helper()
	rl := New(10)
	return rl, HTTPHandler(rl)
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, h := makeHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/runlog", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestHTTPHandler_UnknownJob(t *testing.T) {
	_, h := makeHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/runlog?job=ghost", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHTTPHandler_KnownJob(t *testing.T) {
	rl, h := makeHandler(t)
	now := time.Now()
	rl.Record(Entry{Job: "backup", StartedAt: now, Duration: time.Second, Success: true})

	req := httptest.NewRequest(http.MethodGet, "/runlog?job=backup", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var entry Entry
	if err := json.NewDecoder(rr.Body).Decode(&entry); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if entry.Job != "backup" {
		t.Errorf("expected job 'backup', got %q", entry.Job)
	}
	if !entry.Success {
		t.Error("expected success=true")
	}
}

func TestHTTPHandler_AllJobs(t *testing.T) {
	rl, h := makeHandler(t)
	now := time.Now()
	rl.Record(Entry{Job: "alpha", StartedAt: now, Duration: time.Second, Success: true})
	rl.Record(Entry{Job: "beta", StartedAt: now, Duration: 2 * time.Second, Success: false})

	req := httptest.NewRequest(http.MethodGet, "/runlog", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var entries map[string]Entry
	if err := json.NewDecoder(rr.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestHTTPHandler_ContentType(t *testing.T) {
	_, h := makeHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/runlog", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}
