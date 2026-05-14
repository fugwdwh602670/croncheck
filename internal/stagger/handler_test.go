package stagger

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeHandler() (*Store, http.Handler) {
	s := New()
	return s, HTTPHandler(s)
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/stagger", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_ListEmpty(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stagger", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result []Entry
	_ = json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 0 {
		t.Fatalf("expected empty list, got %d entries", len(result))
	}
}

func TestHTTPHandler_PutAndGet(t *testing.T) {
	_, h := makeHandler()
	body, _ := json.Marshal(map[string]string{"job": "nightly", "delay": "10s"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/stagger", bytes.NewReader(body)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/stagger", nil))
	var result []Entry
	_ = json.NewDecoder(rec2.Body).Decode(&result)
	if len(result) != 1 || result[0].Job != "nightly" || result[0].Delay != 10*time.Second {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestHTTPHandler_PutInvalidDuration(t *testing.T) {
	_, h := makeHandler()
	body, _ := json.Marshal(map[string]string{"job": "x", "delay": "not-a-duration"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/stagger", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHTTPHandler_Delete(t *testing.T) {
	s, h := makeHandler()
	_ = s.Set("weekly", 30*time.Second)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/stagger?job=weekly", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	_, ok := s.Get("weekly")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestHTTPHandler_DeleteMissingParam(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/stagger", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
