package probe

import (
	"encoding/json"
	"net/http"
	"time"
)

type probeResponse struct {
	Job       string `json:"job"`
	Status    string `json:"status"`
	LastProbe string `json:"last_probe,omitempty"`
	TTL       string `json:"ttl,omitempty"`
}

func toResponse(e Entry) probeResponse {
	r := probeResponse{
		Job:    e.Job,
		Status: string(e.Status),
	}
	if !e.LastProbe.IsZero() {
		r.LastProbe = e.LastProbe.UTC().Format(time.RFC3339)
		r.TTL = e.TTL.String()
	}
	return r
}

// HTTPHandler returns an http.Handler for the probe API.
// GET  /probe?job=<name>  — returns status for a single job
// GET  /probe             — returns all probe entries
// POST /probe?job=<name>&ttl=<duration> — records a probe check-in
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			job := r.URL.Query().Get("job")
			w.Header().Set("Content-Type", "application/json")
			if job != "" {
				e, _ := s.Get(job)
				_ = json.NewEncoder(w).Encode(toResponse(e))
				return
			}
			all := s.All()
			resps := make([]probeResponse, len(all))
			for i, e := range all {
				resps[i] = toResponse(e)
			}
			_ = json.NewEncoder(w).Encode(resps)

		case http.MethodPost:
			job := r.URL.Query().Get("job")
			ttlStr := r.URL.Query().Get("ttl")
			if job == "" {
				http.Error(w, "missing job", http.StatusBadRequest)
				return
			}
			ttl, err := time.ParseDuration(ttlStr)
			if err != nil || ttl <= 0 {
				http.Error(w, "invalid ttl", http.StatusBadRequest)
				return
			}
			if err := s.Record(job, ttl); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
