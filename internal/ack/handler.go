package ack

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type addRequest struct {
	Job      string `json:"job"`
	AckedBy  string `json:"acked_by"`
	Duration string `json:"duration"`
}

// HTTPHandler returns an http.Handler for the acknowledgement API.
//
//	GET  /acks          – list active acknowledgements
//	POST /acks          – add an acknowledgement
//	DELETE /acks/{job}  – remove an acknowledgement
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
	_ = json.NewEncoder(w).Encode(s.All())
}

func addAck(w http.ResponseWriter, r *http.Request, s *Store) {
	var req addRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Job == "" {
		http.Error(w, "job is required", http.StatusBadRequest)
		return
	}
	d, err := time.ParseDuration(req.Duration)
	if err != nil || d <= 0 {
		http.Error(w, "invalid duration", http.StatusBadRequest)
		return
	}
	s.Acknowledge(req.Job, req.AckedBy, d)
	w.WriteHeader(http.StatusCreated)
}

func removeAck(w http.ResponseWriter, r *http.Request, s *Store) {
	job := strings.TrimPrefix(r.URL.Path, "/acks/")
	if job == "" {
		http.Error(w, "job is required", http.StatusBadRequest)
		return
	}
	s.Remove(job)
	w.WriteHeader(http.StatusNoContent)
}
