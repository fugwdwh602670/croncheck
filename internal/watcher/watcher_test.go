package watcher_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/croncheck/internal/config"
	"github.com/example/croncheck/internal/notifier"
	"github.com/example/croncheck/internal/scheduler"
	"github.com/example/croncheck/internal/store"
	"github.com/example/croncheck/internal/watcher"
)

func makeWatcher(t *testing.T, webhookURL string, jobs []config.Job, interval time.Duration) *watcher.Watcher {
	t.Helper()
	st := store.New()
	sched := scheduler.New(st, jobs)
	n := notifier.New(webhookURL, 2*time.Second)
	return watcher.New(sched, n, jobs, interval)
}

func TestRun_StopsOnContextCancel(t *testing.T) {
	jobs := []config.Job{{Name: "backup", Schedule: "@hourly", GracePeriod: "5m"}}
	w := makeWatcher(t, "http://localhost:0", jobs, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		w.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
		// expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("watcher did not stop after context cancellation")
	}
}

func TestRun_SendsAlertForMissedJob(t *testing.T) {
	var callCount atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		callCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Grace period already expired — any check should flag this job.
	jobs := []config.Job{{Name: "stale", Schedule: "@hourly", GracePeriod: "0s"}}
	w := makeWatcher(t, ts.URL, jobs, 30*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	w.Run(ctx)

	if callCount.Load() == 0 {
		t.Error("expected at least one alert to be sent for missed job")
	}
}
