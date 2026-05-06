package silence_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/croncheck/internal/silence"
)

func makeHandler() (http.Handler, *silence.Registry) {
	r := silence.New()
	return silence.HTTPHandler(r), r
}

func TestHTTPHandler_ListEmpty(t *testing.T) {
	h, _ := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/silences", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHTTPHandler_AddAndList(t *testing.T) {
	h, reg := makeHandler()
	body, _ := json.Marshal(map[string]string{
		"job_name": "nightly",
		"reason":   "planned outage",
		"duration": "2h",
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/silences", bytes.NewReader(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	if !reg.IsSilenced("nightly", time.Now()) {
		t.Fatal("expected nightly to be silenced after POST")
	}
}

func TestHTTPHandler_AddInvalidDuration(t *testing.T) {
	h, _ := makeHandler()
	body, _ := json.Marshal(map[string]string{"job_name": "x", "duration": "bad"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/silences", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHTTPHandler_Remove(t *testing.T) {
	h, reg := makeHandler()
	reg.Add("cleanup", "test", time.Now().Add(time.Hour))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/silences?job=cleanup", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if reg.IsSilenced("cleanup", time.Now()) {
		t.Fatal("expected silence to be removed")
	}
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	h, _ := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/silences", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
