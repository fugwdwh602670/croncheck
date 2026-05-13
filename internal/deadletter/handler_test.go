package deadletter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeHandler(limit int) (*Store, http.HandlerFunc) {
	s := New(limit)
	return s, HTTPHandler(s)
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	_, h := makeHandler(10)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodPost, "/deadletter", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHTTPHandler_ListEmpty(t *testing.T) {
	_, h := makeHandler(10)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/deadletter", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entries []Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty list, got %d entries", len(entries))
	}
}

func TestHTTPHandler_ListWithEntries(t *testing.T) {
	s, h := makeHandler(10)
	s.Add("nightly", "timeout", `{}`, 2)

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/deadletter", nil))

	var entries []Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Job != "nightly" {
		t.Errorf("job: got %q, want nightly", entries[0].Job)
	}
}

func TestHTTPHandler_RemoveMissingParam(t *testing.T) {
	_, h := makeHandler(10)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodDelete, "/deadletter", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHTTPHandler_RemoveJob(t *testing.T) {
	s, h := makeHandler(10)
	s.Add("nightly", "err", "", 1)
	s.Add("weekly", "err", "", 1)

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodDelete, "/deadletter?job=nightly", nil))

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
	if s.Count() != 1 {
		t.Errorf("expected 1 remaining entry, got %d", s.Count())
	}
}

func TestHTTPHandler_ContentType(t *testing.T) {
	_, h := makeHandler(10)
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/deadletter", nil))
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}
}
