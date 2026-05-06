package ack

import (
	"encoding/json"
	"net/http"
	"time"
)

type addRequest struct {
	JobName  string `json:"job_name"`
	Duration string `json:"duration"`
	Reason   string `json:"reason,omitempty"`
}

// HTTPHandler returns an http.Handler for managing acknowledgements.
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listAcks(w, s)
		case http.MethodPost:
			addAck(w, r, s)
		case http.MethodDelete:
			removeAck(w, r, s)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func listAcks(w http.ResponseWriter, s *Store) {
	w.Header().Set("Content-Type", "application/json")
	all := s.All()
	if all == nil {
		all = []Ack{}
	}
	_ = json.NewEncoder(w).Encode(all)
}

func addAck(w http.ResponseWriter, r *http.Request, s *Store) {
	var req addRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.JobName == "" {
		http.Error(w, "job_name is required", http.StatusBadRequest)
		return
	}
	d, err := time.ParseDuration(req.Duration)
	if err != nil || d <= 0 {
		http.Error(w, "invalid duration", http.StatusBadRequest)
		return
	}
	s.Add(req.JobName, req.Reason, d)
	w.WriteHeader(http.StatusCreated)
}

func removeAck(w http.ResponseWriter, r *http.Request, s *Store) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "job query param required", http.StatusBadRequest)
		return
	}
	s.Remove(job)
	w.WriteHeader(http.StatusNoContent)
}
