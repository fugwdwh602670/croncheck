package ratelimit_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/croncheck/internal/ratelimit"
)

func makeHandler(cooldown time.Duration, jobs []string) http.Handler {
	l := ratelimit.New(cooldown)
	return ratelimit.HTTPHandler(l, jobs)
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	h := makeHandler(time.Minute, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/ratelimit", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHTTPHandler_EmptyJobs(t *testing.T) {
	h := makeHandler(time.Minute, []string{})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ratelimit", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var out []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty array, got %d entries", len(out))
	}
}

func TestHTTPHandler_JobsListed(t *testing.T) {
	jobs := []string{"backup", "report"}
	h := makeHandler(time.Minute, jobs)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ratelimit", nil))

	var out []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
}

func TestHTTPHandler_ContentType(t *testing.T) {
	h := makeHandler(time.Minute, []string{"backup"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ratelimit", nil))
	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}
