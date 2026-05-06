package inhibit

import (
	"encoding/json"
	"net/http"
	"time"
)

type ruleResponse struct {
	SourceJob string `json:"source_job"`
	TargetJob string `json:"target_job"`
}

type statusResponse struct {
	Rules      []ruleResponse       `json:"rules"`
	Unhealthy  map[string]time.Time `json:"unhealthy_sources"`
}

// HTTPHandler returns an http.Handler that exposes inhibition state.
func HTTPHandler(s *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rules := s.Rules()
		resp := statusResponse{
			Rules:     make([]ruleResponse, len(rules)),
			Unhealthy: s.UnhealthySources(),
		}
		for i, r := range rules {
			resp.Rules[i] = ruleResponse{SourceJob: r.SourceJob, TargetJob: r.TargetJob}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}
