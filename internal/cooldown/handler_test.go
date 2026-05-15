package cooldown

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeHandler() (*Store, http.HandlerFunc) {
	s := New()
	return s, HTTPHandler(s)
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPost, "/cooldown", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_ListEmpty(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/cooldown", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var out []entryResponse
	_ = json.NewDecoder(rec.Body).Decode(&out)
	if len(out) != 0 {
		t.Fatalf("want empty list, got %d entries", len(out))
	}
}

func TestHTTPHandler_PutAndGet(t *testing.T) {
	_, h := makeHandler()
	body := bytes.NewBufferString(`{"duration":"2m"}`)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPut, "/cooldown?job=backup", body))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT: want 204, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h(rec2, httptest.NewRequest(http.MethodGet, "/cooldown?job=backup", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("GET: want 200, got %d", rec2.Code)
	}
	var e entryResponse
	_ = json.NewDecoder(rec2.Body).Decode(&e)
	if e.Cooldown != "2m0s" {
		t.Fatalf("want 2m0s, got %s", e.Cooldown)
	}
}

func TestHTTPHandler_PutInvalidDuration(t *testing.T) {
	_, h := makeHandler()
	body := bytes.NewBufferString(`{"duration":"notaduration"}`)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPut, "/cooldown?job=backup", body))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestHTTPHandler_GetUnknownJob(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/cooldown?job=ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestHTTPHandler_DeleteJob(t *testing.T) {
	s, h := makeHandler()
	_ = s.Set("backup", 0+1) // won't error
	_ = s.Set("backup", 60_000_000_000)

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodDelete, "/cooldown?job=backup", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d", rec.Code)
	}
	_, ok := s.Get("backup")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestHTTPHandler_ContentType(t *testing.T) {
	s, h := makeHandler()
	_ = s.Set("backup", 30_000_000_000)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/cooldown", nil))
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("want application/json, got %s", ct)
	}
}
