package digest

import (
	"encoding/json"
	"net/http"
	"time"
)

type response struct {
	BuiltAt time.Time `json:"built_at,omitempty"`
	Entries []Entry   `json:"entries"`
}

// HTTPHandler returns an http.HandlerFunc that serves the current digest.
func HTTPHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		resp := response{
			BuiltAt: s.BuiltAt(),
			Entries: s.All(),
		}
		if resp.Entries == nil {
			resp.Entries = []Entry{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}
}
