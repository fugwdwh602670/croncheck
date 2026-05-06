package history

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HTTPHandler returns an http.Handler that serves job history over HTTP.
func HTTPHandler(h *History) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Expect path: /history/{job}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 || parts[len(parts)-1] == "" {
			http.Error(w, "job name required", http.StatusBadRequest)
			return
		}
		jobName := parts[len(parts)-1]

		events := h.Get(jobName)
		if events == nil {
			events = []Event{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(events); err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
		}
	})
}
