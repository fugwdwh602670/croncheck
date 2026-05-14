package forecast

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makeHandler() (http.HandlerFunc, *Store) {
	s := makeStore()
	return HTTPHandler(s), s
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	h, _ := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/forecast", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestHTTPHandler_UnknownJob(t *testing.T) {
	h, _ := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/forecast?job=ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestHTTPHandler_KnownJob(t *testing.T) {
	h, s := makeHandler()
	base := time.Now()
	_ = s.Record("backup", base)
	_ = s.Record("backup", base.Add(15*time.Minute))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/forecast?job=backup", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var e Entry
	if err := json.NewDecoder(rec.Body).Decode(&e); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if e.Job != "backup" {
		t.Errorf("job = %q, want backup", e.Job)
	}
	if e.AvgInterval != 15*time.Minute {
		t.Errorf("avg interval = %v, want 15m", e.AvgInterval)
	}
}

func TestHTTPHandler_AllJobs(t *testing.T) {
	h, s := makeHandler()
	base := time.Now()
	_ = s.Record("a", base)
	_ = s.Record("a", base.Add(5*time.Minute))
	_ = s.Record("b", base)
	_ = s.Record("b", base.Add(10*time.Minute))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/forecast", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var entries []Entry
	if err := json.NewDecoder(rec.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("len = %d, want 2", len(entries))
	}
}

func TestHTTPHandler_EmptyStore(t *testing.T) {
	h, _ := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/forecast", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var entries []Entry
	_ = json.NewDecoder(rec.Body).Decode(&entries)
	if len(entries) != 0 {
		t.Errorf("expected empty list, got %d entries", len(entries))
	}
}

func TestHTTPHandler_ContentType(t *testing.T) {
	h, _ := makeHandler()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/forecast", nil))
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("content-type = %q, want application/json", ct)
	}
}
