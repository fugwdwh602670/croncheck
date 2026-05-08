package grouping

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeHandler(t *testing.T) (*Store, http.Handler) {
	t.Helper()
	s := New()
	return s, HTTPHandler(s)
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, h := makeHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/grouping", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_PutAndGet(t *testing.T) {
	_, h := makeHandler(t)
	body, _ := json.Marshal(map[string]string{"job": "backup", "group": "ops"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/grouping", bytes.NewReader(body)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/grouping?job=backup", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var resp map[string]string
	_ = json.NewDecoder(rec2.Body).Decode(&resp)
	if resp["group"] != "ops" {
		t.Fatalf("expected ops, got %q", resp["group"])
	}
}

func TestHTTPHandler_GetUnknownJob(t *testing.T) {
	_, h := makeHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/grouping?job=unknown", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHTTPHandler_GetAll(t *testing.T) {
	s, h := makeHandler(t)
	_ = s.Set("backup", "ops")
	_ = s.Set("report", "finance")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/grouping", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp))
	}
}

func TestHTTPHandler_Delete(t *testing.T) {
	s, h := makeHandler(t)
	_ = s.Set("backup", "ops")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/grouping?job=backup", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if _, ok := s.Get("backup"); ok {
		t.Fatal("expected job to be removed")
	}
}

func TestHTTPHandler_PutInvalidJSON(t *testing.T) {
	_, h := makeHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/grouping", bytes.NewBufferString("not-json")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
