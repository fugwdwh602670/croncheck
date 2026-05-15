package expiry

import (
	"encoding/json"
	"net/http"
	"time"
)

type setRequest struct {
	TTL string `json:"ttl"`
}

type entryResponse struct {
	Job      string    `json:"job"`
	Deadline time.Time `json:"deadline"`
	SetAt    time.Time `json:"set_at"`
	Expired  bool      `json:"expired"`
}

func toResponse(s *Store, e Entry) entryResponse {
	return entryResponse{
		Job:      e.Job,
		Deadline: e.Deadline,
		SetAt:    e.SetAt,
		Expired:  s.IsExpired(e.Job),
	}
}

// HTTPHandler handles GET, PUT, and DELETE requests for job expiry deadlines.
func HTTPHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job := r.URL.Query().Get("job")
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			if job == "" {
				all := s.All()
				resp := make([]entryResponse, 0, len(all))
				for _, e := range all {
					resp = append(resp, toResponse(s, e))
				}
				json.NewEncoder(w).Encode(resp)
				return
			}
			e, ok := s.Get(job)
			if !ok {
				http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(toResponse(s, e))

		case http.MethodPut:
			if job == "" {
				http.Error(w, `{"error":"missing job"}`, http.StatusBadRequest)
				return
			}
			var req setRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
				return
			}
			ttl, err := time.ParseDuration(req.TTL)
			if err != nil {
				http.Error(w, `{"error":"invalid ttl"}`, http.StatusBadRequest)
				return
			}
			if err := s.Set(job, ttl); err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case http.MethodDelete:
			if job == "" {
				http.Error(w, `{"error":"missing job"}`, http.StatusBadRequest)
				return
			}
			s.Remove(job)
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	}
}
