package throttle

import (
	"encoding/json"
	"net/http"
	"time"
)

type request struct {
	Job       string `json:"job"`
	MaxAlerts int    `json:"max_alerts"`
	Window    string `json:"window"`
}

type response struct {
	Job       string `json:"job"`
	MaxAlerts int    `json:"max_alerts"`
	Window    string `json:"window"`
}

// HTTPHandler exposes throttle configuration over HTTP.
//
//	GET  /throttle          – list all policies
//	GET  /throttle?job=name – get policy for a job
//	PUT  /throttle          – set policy (JSON body)
//	DELETE /throttle?job=name – remove policy
func HTTPHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			job := r.URL.Query().Get("job")
			if job == "" {
				all := s.All()
				out := make([]response, 0, len(all))
				for j, cfg := range all {
					out = append(out, response{Job: j, MaxAlerts: cfg.MaxAlerts, Window: cfg.Window.String()})
				}
				writeJSON(w, out)
				return
			}
			cfg, ok := s.Get(job)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			writeJSON(w, response{Job: job, MaxAlerts: cfg.MaxAlerts, Window: cfg.Window.String()})

		case http.MethodPut:
			var req request
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			d, err := time.ParseDuration(req.Window)
			if err != nil {
				http.Error(w, "invalid window duration", http.StatusBadRequest)
				return
			}
			if err := s.Set(req.Job, Config{MaxAlerts: req.MaxAlerts, Window: d}); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			job := r.URL.Query().Get("job")
			if job == "" {
				http.Error(w, "missing job", http.StatusBadRequest)
				return
			}
			s.Remove(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
