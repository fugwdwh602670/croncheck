package tags

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HTTPHandler returns an http.Handler that exposes tag CRUD and filtering.
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGet(s, w, r)
		case http.MethodPut:
			handlePut(s, w, r)
		case http.MethodDelete:
			handleDelete(s, w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func handleGet(s *Store, w http.ResponseWriter, r *http.Request) {
	// ?filter=key:value,key2:value2 returns matching jobs
	if f := r.URL.Query().Get("filter"); f != "" {
		query := parseFilter(f)
		jobs := s.Filter(query)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]string{"jobs": jobs})
		return
	}

	// ?job=name returns tags for a specific job
	if job := r.URL.Query().Get("job"); job != "" {
		t, ok := s.Get(job)
		if !ok {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.All())
}

func handlePut(s *Store, w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	var t map[string]string
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := Validate(t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.Set(job, t)
	w.WriteHeader(http.StatusNoContent)
}

func handleDelete(s *Store, w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	s.Remove(job)
	w.WriteHeader(http.StatusNoContent)
}

// parseFilter parses "key:value,key2:value2" into a map.
func parseFilter(f string) map[string]string {
	result := make(map[string]string)
	for _, pair := range strings.Split(f, ",") {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			result[parts[0]] = parts[1]
		}
	}
	return result
}
