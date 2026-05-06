package api

import (
	"net/http"

	"github.com/croncheck/internal/healthz"
	"github.com/croncheck/internal/heartbeat"
	"github.com/croncheck/internal/history"
	"github.com/croncheck/internal/metrics"
	"github.com/croncheck/internal/store"
)

// RouterConfig holds dependencies needed to build the HTTP router.
type RouterConfig struct {
	Store   *store.Store
	History *history.History
	Version string
}

// NewRouter wires all HTTP handlers onto a new ServeMux and returns it.
func NewRouter(cfg RouterConfig) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/healthz", healthz.Handler(cfg.Version))
	mux.Handle("/metrics", metrics.Handler(cfg.Store))
	mux.Handle("/api/v1/jobs", Handler(cfg.Store))
	mux.Handle("/heartbeat", heartbeat.Handler(cfg.Store))
	mux.Handle("/history/", history.HTTPHandler(cfg.History))

	return mux
}
