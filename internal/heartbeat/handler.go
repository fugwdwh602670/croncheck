// Package heartbeat provides an HTTP handler for receiving job heartbeats.
package heartbeat

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/user/croncheck/internal/store"
)

// Recorder is the interface used to record a heartbeat for a named job.
type Recorder interface {
	RecordHeartbeat(name string, t time.Time)
}

// Handler returns an HTTP handler that accepts POST /heartbeat/{job}
// requests and records a heartbeat in the provided store.
func Handler(s Recorder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Expect path: /heartbeat/{job}
		path := strings.TrimPrefix(r.URL.Path, "/heartbeat/")
		jobName := strings.TrimSpace(path)
		if jobName == "" {
			http.Error(w, "job name required", http.StatusBadRequest)
			return
		}

		s.RecordHeartbeat(jobName, time.Now().UTC())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"job":    jobName,
			"status": "ok",
		})
	})
}

// ensure store.Store satisfies Recorder at compile time.
var _ Recorder = (*store.Store)(nil)
