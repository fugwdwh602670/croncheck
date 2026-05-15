package budget

import (
	"encoding/json"
	"net/http"
	"time"
)

type setRequest struct {
	Job    string `json:"job"`
	Limit  int    `json:"limit"`
	Window string `json:"window"`
}

type entryResponse struct {
	Job       string `json:"job"`
	Limit     int    `json:"limit"`
	Window    string `json:"window"`
	Consumed  int    `json:"consumed"`
	WindowEnd string `json:"window_end"`
}

func toResponse(e Entry) entryResponse {
	return entryResponse{
		Job:       e.Job,
		Limit:     e.Limit,
		Window:    e.Window.String(),
		Consumed:  e.Consumed,
		WindowEnd: e.WindowEnd.UTC().Format(time.RFC3339),
	}
}

// HTTPHandler returns an http.Handler for the budget API.
//
//	GET  /budget?job=<name>  — get entry (omit job for all)
//	PUT  /budget             — set budget {job, limit, window}
//	DELETE /budget?job=<name> — remove entry
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			job := r.URL.Query().Get("job")
			if job != "" {
				e, ok := s.Get(job)
				if !ok {
					http.Error(w, "not found", http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(toResponse(e))
				return
			}
			all := s.All()
			out := make([]entryResponse, len(all))
			for i, e := range all {
				out[i] = toResponse(e)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(out)

		case http.MethodPut:
			var req setRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			d, err := time.ParseDuration(req.Window)
			if err != nil {
				http.Error(w, "invalid window duration", http.StatusBadRequest)
				return
			}
			if err := s.Set(req.Job, req.Limit, d); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			job := r.URL.Query().Get("job")
			if job == "" {
				http.Error(w, "job param required", http.StatusBadRequest)
				return
			}
			s.Remove(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
