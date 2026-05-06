package retention

import (
	"encoding/json"
	"net/http"
	"time"
)

// StatusReporter can report the configured retention TTL and interval.
type StatusReporter struct {
	TTL      time.Duration
	Interval time.Duration
}

type statusResponse struct {
	TTL      string `json:"ttl"`
	Interval string `json:"interval"`
}

// HTTPHandler returns an http.HandlerFunc that reports retention configuration.
func HTTPHandler(sr StatusReporter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		resp := statusResponse{
			TTL:      sr.TTL.String(),
			Interval: sr.Interval.String(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}
}
