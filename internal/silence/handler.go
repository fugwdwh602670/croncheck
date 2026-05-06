package silence

import (
	"encoding/json"
	"net/http"
	"time"
)

type addRequest struct {
	JobName  string `json:"job_name"`
	Reason   string `json:"reason"`
	Duration string `json:"duration"` // e.g. "2h", "30m"
}

// HTTPHandler returns an http.Handler that exposes silence management endpoints.
//
//	POST   /silences        – add a silence
//	DELETE /silences?job=…  – remove a silence
//	GET    /silences        – list active silences
func HTTPHandler(r *Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			listSilences(w, r)
		case http.MethodPost:
			addSilence(w, req, r)
		case http.MethodDelete:
			removeSilence(w, req, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func listSilences(w http.ResponseWriter, r *Registry) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(r.All(time.Now()))
}

func addSilence(w http.ResponseWriter, req *http.Request, r *Registry) {
	var body addRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil || body.JobName == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	d, err := time.ParseDuration(body.Duration)
	if err != nil || d <= 0 {
		http.Error(w, "invalid duration", http.StatusBadRequest)
		return
	}
	r.Add(body.JobName, body.Reason, time.Now().Add(d))
	w.WriteHeader(http.StatusCreated)
}

func removeSilence(w http.ResponseWriter, req *http.Request, r *Registry) {
	job := req.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job query parameter", http.StatusBadRequest)
		return
	}
	r.Remove(job)
	w.WriteHeader(http.StatusNoContent)
}
