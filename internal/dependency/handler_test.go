package dependency

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeHandler(t *testing.T) (http.Handler, *Store) {
	t.Helper()
	s := New()
	return HTTPHandler(s), s
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	h, _ := makeHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/dependencies", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_PutAndGetJob(t *testing.T) {
	h, _ := makeHandler(t)
	body, _ := json.Marshal(map[string][]string{"upstreams": {"job-a"}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/dependencies?job=job-b", bytes.NewReader(body)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/dependencies?job=job-b", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var resp map[string][]string
	_ = json.NewDecoder(rec2.Body).Decode(&resp)
	if len(resp["upstreams"]) != 1 || resp["upstreams"][0] != "job-a" {
		t.Fatalf("unexpected upstreams: %v", resp)
	}
}

func TestHTTPHandler_GetAll(t *testing.T) {
	h, s := makeHandler(t)
	_ = s.Set("job-b", []string{"job-a"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dependencies", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var all map[string][]string
	_ = json.NewDecoder(rec.Body).Decode(&all)
	if _, ok := all["job-b"]; !ok {
		t.Fatal("expected job-b in response")
	}
}

func TestHTTPHandler_DeleteJob(t *testing.T) {
	h, s := makeHandler(t)
	_ = s.Set("job-b", []string{"job-a"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/dependencies?job=job-b", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if ups := s.Get("job-b"); ups != nil {
		t.Fatalf("expected nil after delete, got %v", ups)
	}
}

func TestHTTPHandler_PutSelfDependencyReturns400(t *testing.T) {
	h, _ := makeHandler(t)
	body, _ := json.Marshal(map[string][]string{"upstreams": {"job-a"}})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/dependencies?job=job-a", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
