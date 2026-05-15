package expiry

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
	h(rec, httptest.NewRequest(http.MethodPost, "/expiry", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_PutAndGet(t *testing.T) {
	_, h := makeHandler()

	body := bytes.NewBufferString(`{"ttl":"10m"}`)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPut, "/expiry?job=backup", body))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT: expected 204, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/expiry?job=backup", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET: expected 200, got %d", rec.Code)
	}
	var resp entryResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Job != "backup" {
		t.Errorf("job = %q, want backup", resp.Job)
	}
}

func TestHTTPHandler_GetUnknownJob(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/expiry?job=ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHTTPHandler_GetAll(t *testing.T) {
	_, h := makeHandler()
	for _, job := range []string{"a", "b"} {
		body := bytes.NewBufferString(`{"ttl":"5m"}`)
		h(httptest.NewRecorder(), httptest.NewRequest(http.MethodPut, "/expiry?job="+job, body))
	}
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/expiry", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp []entryResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 entries, got %d", len(resp))
	}
}

func TestHTTPHandler_PutInvalidTTL(t *testing.T) {
	_, h := makeHandler()
	body := bytes.NewBufferString(`{"ttl":"not-a-duration"}`)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPut, "/expiry?job=backup", body))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHTTPHandler_DeleteJob(t *testing.T) {
	_, h := makeHandler()
	body := bytes.NewBufferString(`{"ttl":"5m"}`)
	h(httptest.NewRecorder(), httptest.NewRequest(http.MethodPut, "/expiry?job=backup", body))

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodDelete, "/expiry?job=backup", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/expiry?job=backup", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rec.Code)
	}
}
