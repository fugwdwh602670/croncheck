// Package healthz provides a simple HTTP health check endpoint.
package healthz

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response is the JSON body returned by the health endpoint.
type Response struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
}

// Handler returns an http.HandlerFunc that responds to health check requests.
// It always returns HTTP 200 with a JSON body as long as the process is alive.
func Handler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		resp := Response{
			Status:    "ok",
			Timestamp: time.Now().UTC(),
			Version:   version,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
}
