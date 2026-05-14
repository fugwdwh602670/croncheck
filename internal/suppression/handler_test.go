package suppression

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeHandler(t *testing.T) (*Store, http.HandlerFunc) {
	t.Helper()
	s := New()
	return s, HTTPHandler(s)
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, h := makeHandler(t)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPost, "/suppression", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_ListEmpty(t *testing.T) {
	_, h := makeHandler(t)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/suppression", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var rules []Rule
	if err := json.NewDecoder(rec.Body).Decode(&rules); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(rules) != 0 {
		t.Errorf("expected empty list")
	}
}

func TestHTTPHandler_PutAndGet(t *testing.T) {
	_, h := makeHandler(t)

	// PUT
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPut, "/suppression?job=backup&min_consec_misses=3", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	// GET single
	rec = httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/suppression?job=backup", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var rule Rule
	if err := json.NewDecoder(rec.Body).Decode(&rule); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if rule.MinConsecMisses != 3 {
		t.Errorf("expected 3, got %d", rule.MinConsecMisses)
	}
}

func TestHTTPHandler_PutInvalidThreshold(t *testing.T) {
	_, h := makeHandler(t)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPut, "/suppression?job=backup&min_consec_misses=0", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHTTPHandler_DeleteJob(t *testing.T) {
	s, h := makeHandler(t)
	_ = s.Set("backup", 2)

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodDelete, "/suppression?job=backup", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	_, ok := s.Get("backup")
	if ok {
		t.Error("expected rule to be deleted")
	}
}

func TestHTTPHandler_GetUnknownJob(t *testing.T) {
	_, h := makeHandler(t)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/suppression?job=ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
