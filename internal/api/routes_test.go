package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/croncheck/internal/history"
	"github.com/croncheck/internal/store"
)

func makeRouter(t *testing.T) *http.ServeMux {
	t.Helper()
	s := store.New()
	h := history.New(50)
	return NewRouter(RouterConfig{
		Store:   s,
		History: h,
		Version: "test-v1",
	})
}

func TestRoutes_Healthz(t *testing.T) {
	mux := makeRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/healthz: expected 200, got %d", rec.Code)
	}
}

func TestRoutes_Metrics(t *testing.T) {
	mux := makeRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/metrics: expected 200, got %d", rec.Code)
	}
}

func TestRoutes_Jobs(t *testing.T) {
	mux := makeRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jobs", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/api/v1/jobs: expected 200, got %d", rec.Code)
	}
}

func TestRoutes_History(t *testing.T) {
	mux := makeRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/history/myjob", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("/history/myjob: expected 200, got %d", rec.Code)
	}
}

func TestRoutes_HeartbeatMethodNotAllowed(t *testing.T) {
	mux := makeRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/heartbeat", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("/heartbeat GET: expected 405, got %d", rec.Code)
	}
}
