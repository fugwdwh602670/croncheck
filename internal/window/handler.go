package window

import (
	"encoding/json"
	"net/http"
	"time"
)

type putRequest struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type entryResponse struct {
	Job       string    `json:"job"`
	Start     string    `json:"start"`
	End       string    `json:"end"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toResponse(e Entry) entryResponse {
	return entryResponse{
		Job:       e.Job,
		Start:     e.Start.String(),
		End:       e.End.String(),
		UpdatedAt: e.UpdatedAt,
	}
}

// HTTPHandler returns an http.Handler for the window store.
// GET  /window?job=<name>  — retrieve window for a job (omit job for all)
// PUT  /window?job=<name>  — set window {"start":"0s","end":"5m"}
// DELETE /window?job=<name> — remove window
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		job := r.URL.Query().Get("job")
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			if job == "" {
				json.NewEncoder(w).Encode(func() []entryResponse {
					all := s.All()
					out := make([]entryResponse, len(all))
					for i, e := range all {
						out[i] = toResponse(e)
					}
					return out
				}())
				return
			}
			e, ok := s.Get(job)
			if !ok {
				http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(toResponse(e))

		case http.MethodPut:
			if job == "" {
				http.Error(w, `{"error":"job required"}`, http.StatusBadRequest)
				return
			}
			var req putRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
				return
			}
			start, err := time.ParseDuration(req.Start)
			if err != nil {
				http.Error(w, `{"error":"invalid start duration"}`, http.StatusBadRequest)
				return
			}
			end, err := time.ParseDuration(req.End)
			if err != nil {
				http.Error(w, `{"error":"invalid end duration"}`, http.StatusBadRequest)
				return
			}
			if err := s.Set(job, start, end); err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if job == "" {
				http.Error(w, `{"error":"job required"}`, http.StatusBadRequest)
				return
			}
			s.Remove(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})
}
