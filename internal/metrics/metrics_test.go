package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/croncheck/internal/metrics"
	"github.com/example/croncheck/internal/store"
)

type fakeProvider struct {
	jobs []store.JobState
}

func (f *fakeProvider) All() []store.JobState { return f.jobs }

func TestMetrics_HealthyJob(t *testing.T) {
	p := &fakeProvider{
		jobs: []store.JobState{
			{Name: "backup", LastSeen: time.Now().Add(-30 * time.Second), MissedCount: 0},
		},
	}

	rec := httptest.NewRecorder()
	metrics.Handler(p).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := rec.Body.String()
	if !strings.Contains(body, `croncheck_job_healthy{job="backup"} 1`) {
		t.Errorf("expected healthy=1, got:\n%s", body)
	}
	if !strings.Contains(body, `croncheck_job_missed_total{job="backup"} 0`) {
		t.Errorf("expected missed=0, got:\n%s", body)
	}
}

func TestMetrics_NeverSeenJob(t *testing.T) {
	p := &fakeProvider{
		jobs: []store.JobState{
			{Name: "report", LastSeen: time.Time{}, MissedCount: 3},
		},
	}

	rec := httptest.NewRecorder()
	metrics.Handler(p).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := rec.Body.String()
	if !strings.Contains(body, `croncheck_job_healthy{job="report"} 0`) {
		t.Errorf("expected healthy=0, got:\n%s", body)
	}
	if !strings.Contains(body, `croncheck_job_last_seen_seconds{job="report"} -1.000`) {
		t.Errorf("expected last_seen=-1, got:\n%s", body)
	}
	if !strings.Contains(body, `croncheck_job_missed_total{job="report"} 3`) {
		t.Errorf("expected missed=3, got:\n%s", body)
	}
}

func TestMetrics_ContentType(t *testing.T) {
	p := &fakeProvider{}
	rec := httptest.NewRecorder()
	metrics.Handler(p).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("unexpected Content-Type: %s", ct)
	}
}

func TestSanitizeJobName(t *testing.T) {
	p := &fakeProvider{
		jobs: []store.JobState{
			{Name: "my job@home", LastSeen: time.Now(), MissedCount: 0},
		},
	}

	rec := httptest.NewRecorder()
	metrics.Handler(p).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := rec.Body.String()
	if strings.Contains(body, "@") {
		t.Errorf("expected @ to be sanitized, got:\n%s", body)
	}
}
