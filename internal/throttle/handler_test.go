package throttle

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
	h(rec, httptest.NewRequest(http.MethodPost, "/throttle", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_ListEmpty(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/throttle", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out []response
	_ = json.NewDecoder(rec.Body).Decode(&out)
	if len(out) != 0 {
		t.Fatalf("expected empty list, got %d", len(out))
	}
}

func TestHTTPHandler_PutAndGet(t *testing.T) {
	_, h := makeHandler()
	body, _ := json.Marshal(request{Job: "backup", MaxAlerts: 3, Window: "5m"})
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPut, "/throttle", bytes.NewReader(body)))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h(rec2, httptest.NewRequest(http.MethodGet, "/throttle?job=backup", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var resp response
	_ = json.NewDecoder(rec2.Body).Decode(&resp)
	if resp.MaxAlerts != 3 {
		t.Errorf("expected MaxAlerts=3, got %d", resp.MaxAlerts)
	}
	if resp.Window != "5m0s" {
		t.Errorf("expected window=5m0s, got %s", resp.Window)
	}
}

func TestHTTPHandler_GetUnknownJob(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/throttle?job=ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHTTPHandler_PutInvalidDuration(t *testing.T) {
	_, h := makeHandler()
	body, _ := json.Marshal(request{Job: "job", MaxAlerts: 1, Window: "bad"})
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPut, "/throttle", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHTTPHandler_Delete(t *testing.T) {
	s, h := makeHandler()
	_ = s.Set("job", Config{MaxAlerts: 1, Window: 60 * 1000000000})

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodDelete, "/throttle?job=job", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	_, ok := s.Get("job")
	if ok {
		t.Error("expected policy to be removed")
	}
}
