package tags

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeHandler() (*Store, http.Handler) {
	s := New()
	return s, HTTPHandler(s)
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/tags", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_PutAndGetJob(t *testing.T) {
	_, h := makeHandler()
	body, _ := json.Marshal(map[string]string{"env": "prod", "team": "ops"})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/tags?job=backup", bytes.NewReader(body)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tags?job=backup", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got map[string]string
	json.NewDecoder(rec.Body).Decode(&got)
	if got["env"] != "prod" {
		t.Errorf("unexpected tags: %v", got)
	}
}

func TestHTTPHandler_GetUnknownJob(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tags?job=ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHTTPHandler_Filter(t *testing.T) {
	s, h := makeHandler()
	s.Set("job-a", map[string]string{"env": "prod"})
	s.Set("job-b", map[string]string{"env": "staging"})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/tags?filter=env:prod", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string][]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp["jobs"]) != 1 || resp["jobs"][0] != "job-a" {
		t.Errorf("unexpected filter result: %v", resp)
	}
}

func TestHTTPHandler_Delete(t *testing.T) {
	s, h := makeHandler()
	s.Set("job-a", map[string]string{"env": "prod"})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/tags?job=job-a", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	_, ok := s.Get("job-a")
	if ok {
		t.Error("expected tags to be deleted")
	}
}

func TestHTTPHandler_PutMissingJob(t *testing.T) {
	_, h := makeHandler()
	body, _ := json.Marshal(map[string]string{"env": "prod"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/tags", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
