package catalog

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeHandler(t *testing.T) (http.HandlerFunc, *Store) {
	t.Helper()
	s := New()
	return HTTPHandler(s), s
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	h, _ := makeHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/catalog", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_PutAndGet(t *testing.T) {
	h, _ := makeHandler(t)
	body, _ := json.Marshal(Entry{Job: "sync", Owner: "team-b", Schedule: "@hourly"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/catalog", bytes.NewReader(body)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/catalog?job=sync", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var got Entry
	_ = json.NewDecoder(rec2.Body).Decode(&got)
	if got.Owner != "team-b" {
		t.Errorf("expected team-b, got %s", got.Owner)
	}
}

func TestHTTPHandler_GetUnknownJob(t *testing.T) {
	h, _ := makeHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/catalog?job=ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHTTPHandler_GetAll(t *testing.T) {
	h, s := makeHandler(t)
	_ = s.Set(Entry{Job: "a"})
	_ = s.Set(Entry{Job: "b"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/catalog", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []Entry
	_ = json.NewDecoder(rec.Body).Decode(&entries)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestHTTPHandler_DeleteJob(t *testing.T) {
	h, s := makeHandler(t)
	_ = s.Set(Entry{Job: "temp"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/catalog?job=temp", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if _, ok := s.Get("temp"); ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestHTTPHandler_PutInvalidJSON(t *testing.T) {
	h, _ := makeHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/catalog", bytes.NewBufferString("not-json")))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
