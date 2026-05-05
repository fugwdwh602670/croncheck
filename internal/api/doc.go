// Package api exposes a lightweight HTTP status endpoint for croncheck.
//
// The single GET /status handler returns a JSON array of all tracked jobs,
// including their last-seen timestamp, missed-run count and current state
// (healthy / missed / unknown). It is intended for health-check dashboards
// and simple polling integrations.
package api
