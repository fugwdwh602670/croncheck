package grouping

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HTTPHandler returns an http.Handler for the grouping API.
//
//	GET    /grouping?job=<name>  – get group for a single job
//	GET    /grouping             – list all job→group mappings
//	PUT    /grouping             – assign {"job":"…","group":"…"}
//	DELETE /grouping?job=<name>  – remove group assignment
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGet(w, r, s)
		case http.MethodPut:
			handlePut(w, r, s)
		case http.MethodDelete:
			handleDelete(w, r, s)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGet(w http.ResponseWriter, r *http.Request, s *Store) {
	w.Header().Set("Content-Type", "application/json")
	if job := strings.TrimSpace(r.URL.Query().Get("job")); job != "" {
		g, ok := s.Get(job)
		if !ok {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"job": job, "group": g})
		return
	}
	_ = json.NewEncoder(w).Encode(s.All())
}

func handlePut(w http.ResponseWriter, r *http.Request, s *Store) {
	var body struct {
		Job   string `json:"job"`
		Group string `json:"group"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := s.Set(body.Job, body.Group); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleDelete(w http.ResponseWriter, r *http.Request, s *Store) {
	job := strings.TrimSpace(r.URL.Query().Get("job"))
	if job == "" {
		http.Error(w, "job query param required", http.StatusBadRequest)
		return
	}
	s.Remove(job)
	w.WriteHeader(http.StatusNoContent)
}
