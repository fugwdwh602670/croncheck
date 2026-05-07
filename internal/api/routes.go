package api

import (
	"net/http"

	"croncheck/internal/ack"
	"croncheck/internal/annotations"
	"croncheck/internal/dashboard"
	"croncheck/internal/dependency"
	"croncheck/internal/heartbeat"
	"croncheck/internal/history"
	"croncheck/internal/inhibit"
	"croncheck/internal/metrics"
	"croncheck/internal/oncall"
	"croncheck/internal/ratelimit"
	"croncheck/internal/retention"
	"croncheck/internal/runlog"
	"croncheck/internal/silence"
	"croncheck/internal/store"
	"croncheck/internal/tags"
	"croncheck/internal/healthz"
)

// RouterConfig holds all dependencies needed to wire up HTTP routes.
type RouterConfig struct {
	Version    string
	Store      *store.Store
	History    *history.Store
	Silence    *silence.Store
	Ack        *ack.Store
	Inhibit    *inhibit.Store
	RateLimit  *ratelimit.Store
	Retention  *retention.Runner
	RunLog     *runlog.Store
	Tags       *tags.Store
	Annotations *annotations.Store
	Dependency *dependency.Store
	Oncall     *oncall.Store
}

// NewRouter constructs and returns a fully wired ServeMux.
func NewRouter(cfg RouterConfig) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/healthz", healthz.Handler(cfg.Version))
	mux.Handle("/metrics", metrics.Handler(cfg.Store))
	mux.Handle("/heartbeat", heartbeat.Handler(cfg.Store))
	mux.Handle("/api/v1/jobs", Handler(cfg.Store))
	mux.Handle("/api/v1/history", history.HTTPHandler(cfg.History))
	mux.Handle("/api/v1/silence", silence.HTTPHandler(cfg.Silence))
	mux.Handle("/api/v1/ack", ack.HTTPHandler(cfg.Ack))
	mux.Handle("/api/v1/inhibit", inhibit.HTTPHandler(cfg.Inhibit))
	mux.Handle("/api/v1/ratelimit", ratelimit.HTTPHandler(cfg.RateLimit))
	mux.Handle("/api/v1/retention", retention.HTTPHandler(cfg.Retention))
	mux.Handle("/api/v1/runlog", runlog.HTTPHandler(cfg.RunLog))
	mux.Handle("/api/v1/tags", tags.HTTPHandler(cfg.Tags))
	mux.Handle("/api/v1/annotations", annotations.HTTPHandler(cfg.Annotations))
	mux.Handle("/api/v1/dependencies", dependency.HTTPHandler(cfg.Dependency))
	mux.Handle("/api/v1/oncall", oncall.HTTPHandler(cfg.Oncall))
	mux.Handle("/api/v1/dashboard", dashboard.HTTPHandler(cfg.Store))

	return mux
}
