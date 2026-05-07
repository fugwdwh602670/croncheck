package oncall_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"croncheck/internal/oncall"
)

func makeHandler(t *testing.T) http.Handler {
	t.Helper()
	return oncall.HTTPHandler(oncall.New())
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	h := makeHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/oncall?job=myjob", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_MissingJobName(t *testing.T) {
	h := makeHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/oncall", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHTTPHandler_PutAndGet(t *testing.T) {
	h := makeHandler(t)

	body := map[string]string{
		"assignee": "alice",
		"ends_at":  time.Now().Add(2 * time.Hour).Format(time.RFC3339),
	}
	b, _ := json.Marshal(body)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/oncall?job=backup", bytes.NewReader(b))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/oncall?job=backup", nil)
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var entry map[string]string
	if err := json.NewDecoder(rec2.Body).Decode(&entry); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if entry["assignee"] != "alice" {
		t.Errorf("expected assignee alice, got %q", entry["assignee"])
	}
}

func TestHTTPHandler_GetUnknownJob(t *testing.T) {
	h := makeHandler(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/oncall?job=ghost", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHTTPHandler_Delete(t *testing.T) {
	store := oncall.New()
	h := oncall.HTTPHandler(store)

	_ = store.Set("backup", oncall.Entry{
		Assignee: "bob",
		EndsAt:   time.Now().Add(time.Hour),
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/oncall?job=backup", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	_, err := store.Get("backup")
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestHTTPHandler_PutExpiredEntry(t *testing.T) {
	h := makeHandler(t)
	body := map[string]string{
		"assignee": "carol",
		"ends_at":  time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
	}
	b, _ := json.Marshal(body)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/oncall?job=sync", bytes.NewReader(b))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for past ends_at, got %d", rec.Code)
	}
}
