package annotations

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
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/annotations", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_PutAndGetJob(t *testing.T) {
	h, _ := makeHandler(t)
	body, _ := json.Marshal(map[string]string{"owner": "ops"})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/annotations?job=backup", bytes.NewReader(body)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/annotations?job=backup", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if got["owner"] != "ops" {
		t.Fatalf("unexpected response: %v", got)
	}
}

func TestHTTPHandler_GetUnknownJob(t *testing.T) {
	h, _ := makeHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/annotations?job=ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHTTPHandler_GetAll(t *testing.T) {
	h, s := makeHandler(t)
	s.Set("j1", map[string]string{"a": "1"})
	s.Set("j2", map[string]string{"b": "2"})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/annotations", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got map[string]map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if len(got) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(got))
	}
}

func TestHTTPHandler_Delete(t *testing.T) {
	h, s := makeHandler(t)
	s.Set("job", map[string]string{"x": "y"})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/annotations?job=job", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	_, ok := s.Get("job")
	if ok {
		t.Fatal("expected annotation to be deleted")
	}
}

func TestHTTPHandler_PutInvalidJSON(t *testing.T) {
	h, _ := makeHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/annotations?job=x", bytes.NewBufferString("not-json")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
