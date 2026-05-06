package heartbeat_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/user/croncheck/internal/heartbeat"
)

// fakeRecorder captures RecordHeartbeat calls for assertions.
type fakeRecorder struct {
	mu   sync.Mutex
	calls []struct {
		name string
		t    time.Time
	}
}

func (f *fakeRecorder) RecordHeartbeat(name string, t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, struct {
		name string
		t    time.Time
	}{name, t})
}

func TestHandler_RecordsHeartbeat(t *testing.T) {
	rec := &fakeRecorder{}
	h := heartbeat.Handler(rec)

	req := httptest.NewRequest(http.MethodPost, "/heartbeat/backup-job", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if len(rec.calls) != 1 || rec.calls[0].name != "backup-job" {
		t.Fatalf("expected heartbeat for 'backup-job', got %+v", rec.calls)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["job"] != "backup-job" || body["status"] != "ok" {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestHandler_MissingJobName(t *testing.T) {
	rec := &fakeRecorder{}
	h := heartbeat.Handler(rec)

	req := httptest.NewRequest(http.MethodPost, "/heartbeat/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if len(rec.calls) != 0 {
		t.Fatal("expected no heartbeat calls")
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	rec := &fakeRecorder{}
	h := heartbeat.Handler(rec)

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		req := httptest.NewRequest(method, "/heartbeat/some-job", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s: expected 405, got %d", method, w.Code)
		}
	}
}
