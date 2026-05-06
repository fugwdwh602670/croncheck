package runlog

import (
	"encoding/json"
	"net/http"
)

// HTTPHandler returns an http.Handler that exposes the run log over HTTP.
//
// GET /runlog?job=<name> returns the latest run entry for the given job.
// GET /runlog returns all known run entries.
func HTTPHandler(rl *RunLog) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		job := r.URL.Query().Get("job")
		if job != "" {
			entry, ok := rl.Get(job)
			if !ok {
				http.Error(w, "job not found", http.StatusNotFound)
				return
			}
			if err := json.NewEncoder(w).Encode(entry); err != nil {
				http.Error(w, "encoding error", http.StatusInternalServerError)
			}
			return
		}

		all := rl.All()
		if err := json.NewEncoder(w).Encode(all); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	})
}
