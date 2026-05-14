package snapshot

import (
	"encoding/json"
	"net/http"
	"strings"
)

// HTTPHandler returns an http.Handler for the snapshot API.
// GET  /snapshots          → list all snapshots (newest first)
// POST /snapshots?id=<id>  → capture a new snapshot
// GET  /snapshots/<id>     → retrieve a specific snapshot
func HTTPHandler(store *Store, src Source) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/snapshots")
		path = strings.TrimPrefix(path, "/")

		switch r.Method {
		case http.MethodGet:
			if path == "" {
				all := store.All()
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(all)
				return
			}
			snap, ok := store.Get(path)
			if !ok {
				http.Error(w, "snapshot not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(snap)

		case http.MethodPost:
			id := r.URL.Query().Get("id")
			if id == "" {
				http.Error(w, "missing id parameter", http.StatusBadRequest)
				return
			}
			snap := store.Capture(id, src)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(snap)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
