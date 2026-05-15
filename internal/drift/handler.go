package drift

import (
	"encoding/json"
	"net/http"
	"time"
)

type driftResponse struct {
	Job         string  `json:"job"`
	LastDriftMs float64 `json:"last_drift_ms"`
	MaxDriftMs  float64 `json:"max_drift_ms"`
	AvgDriftMs  float64 `json:"avg_drift_ms"`
	Samples     int     `json:"sample_count"`
	RecordedAt  string  `json:"recorded_at"`
}

func toResponse(e Entry) driftResponse {
	return driftResponse{
		Job:         e.Job,
		LastDriftMs: float64(e.LastDrift) / float64(time.Millisecond),
		MaxDriftMs:  float64(e.MaxDrift) / float64(time.Millisecond),
		AvgDriftMs:  float64(e.AvgDrift) / float64(time.Millisecond),
		Samples:     e.SampleCount,
		RecordedAt:  e.RecordedAt.UTC().Format(time.RFC3339),
	}
}

// HTTPHandler serves drift statistics.
//
//	GET /drift?job=<name>  — single job
//	GET /drift             — all jobs
func HTTPHandler(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if job := r.URL.Query().Get("job"); job != "" {
			e, ok := s.Get(job)
			if !ok {
				http.Error(w, "job not found", http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(toResponse(e))
			return
		}

		all := s.All()
		out := make([]driftResponse, 0, len(all))
		for _, e := range all {
			out = append(out, toResponse(e))
		}
		json.NewEncoder(w).Encode(out)
	}
}
