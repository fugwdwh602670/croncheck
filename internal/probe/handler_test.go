package probe

import (
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
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/probe", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_PostAndGetJob(t *testing.T) {
	_, h := makeHandler()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/probe?job=backup&ttl=5m", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/probe?job=backup", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp probeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Job != "backup" {
		t.Errorf("expected job=backup, got %s", resp.Job)
	}
	if resp.Status != string(StatusAlive) {
		t.Errorf("expected alive, got %s", resp.Status)
	}
}

func TestHTTPHandler_PostMissingJob(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/probe?ttl=5m", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHTTPHandler_PostInvalidTTL(t *testing.T) {
	_, h := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/probe?job=sync&ttl=bad", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHTTPHandler_GetAll(t *testing.T) {
	_, h := makeHandler()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/probe?job=a&ttl=1m", nil))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/probe?job=b&ttl=2m", nil))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/probe", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resps []probeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resps); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resps) != 2 {
		t.Errorf("expected 2 entries, got %d", len(resps))
	}
}
