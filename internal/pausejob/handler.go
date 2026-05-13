package pausejob

import (
	"encoding/json"
	"net/http"
	"time"
)

type pauseRequest struct {
	Job      string `json:"job"`
	Reason   string `json:"reason"`
	Duration string `json:"duration"`
}

// HTTPHandler returns an http.HandlerFunc that exposes the pause store.
//
//	GET  /pausejob          — list active pauses
//	POST /pausejob          — add a pause  {job, reason, duration}
//	DELETE /pausejob?job=X  — resume a job
func HTTPHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listPauses(w, s)
		case http.MethodPost:
			addPause(w, r, s)
		case http.MethodDelete:
			removePause(w, r, s)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func listPauses(w http.ResponseWriter, s *Store) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.All())
}

func addPause(w http.ResponseWriter, r *http.Request, s *Store) {
	var req pauseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	d, err := time.ParseDuration(req.Duration)
	if err != nil {
		http.Error(w, "invalid duration", http.StatusBadRequest)
		return
	}
	if err := s.Pause(req.Job, req.Reason, d); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func removePause(w http.ResponseWriter, r *http.Request, s *Store) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	s.Resume(job)
	w.WriteHeader(http.StatusNoContent)
}
