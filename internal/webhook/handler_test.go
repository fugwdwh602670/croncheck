package webhook

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
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/webhooks", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestHTTPHandler_PutAndGet(t *testing.T) {
	h, _ := makeHandler(t)
	body, _ := json.Marshal(map[string]string{"job": "backup", "url": "https://hook.example.com", "secret": "s"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/webhooks", bytes.NewReader(body)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("put status = %d, want 204", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/webhooks?job=backup", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("get status = %d, want 200", rec2.Code)
	}
	var e Entry
	if err := json.NewDecoder(rec2.Body).Decode(&e); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if e.URL != "https://hook.example.com" {
		t.Errorf("url = %q", e.URL)
	}
}

func TestHTTPHandler_GetUnknownJob(t *testing.T) {
	h, _ := makeHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/webhooks?job=ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestHTTPHandler_GetAll(t *testing.T) {
	h, s := makeHandler(t)
	_ = s.Set("job-a", "https://a.example.com", "")
	_ = s.Set("job-b", "https://b.example.com", "")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/webhooks", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var entries []Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("len = %d, want 2", len(entries))
	}
}

func TestHTTPHandler_Delete(t *testing.T) {
	h, s := makeHandler(t)
	_ = s.Set("backup", "https://hook.example.com", "")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/webhooks?job=backup", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204", rec.Code)
	}
	_, ok := s.Get("backup")
	if ok {
		t.Error("expected entry to be deleted")
	}
}
