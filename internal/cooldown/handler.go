package cooldown

import (
	"encoding/json"
	"net/http"
	"time"
)

type entryResponse struct {
	Job      string        `json:"job"`
	Cooldown string        `json:"cooldown"`
	LastSeen *time.Time    `json:"last_seen,omitempty"`
}

// HTTPHandler exposes cooldown configuration over HTTP.
//
//	GET  /cooldown          – list all configured cooldowns
//	GET  /cooldown?job=X    – get cooldown for a specific job
//	PUT  /cooldown?job=X    – set cooldown (body: {"duration":"1m"})
//	DELETE /cooldown?job=X  – remove cooldown
func HTTPHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job := r.URL.Query().Get("job")
		switch r.Method {
		case http.MethodGet:
			if job == "" {
				all := s.All()
				out := make([]entryResponse, 0, len(all))
				for name, e := range all {
					out = append(out, toResponse(name, e))
				}
				writeJSON(w, out)
				return
			}
			e, ok := s.Get(job)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			writeJSON(w, toResponse(job, e))

		case http.MethodPut:
			if job == "" {
				http.Error(w, "job query param required", http.StatusBadRequest)
				return
			}
			var body struct {
				Duration string `json:"duration"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			d, err := time.ParseDuration(body.Duration)
			if err != nil {
				http.Error(w, "invalid duration", http.StatusBadRequest)
				return
			}
			if err := s.Set(job, d); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if job == "" {
				http.Error(w, "job query param required", http.StatusBadRequest)
				return
			}
			s.Remove(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func toResponse(job string, e Entry) entryResponse {
	r := entryResponse{Job: job, Cooldown: e.Duration.String()}
	if !e.LastSeen.IsZero() {
		t := e.LastSeen
		r.LastSeen = &t
	}
	return r
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
