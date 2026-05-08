// Package api exposes a lightweight HTTP status endpoint for croncheck.
//
// The single GET /status handler returns a JSON array of all tracked jobs,
// including their last-seen timestamp, missed-run count and current state
// (healthy / missed / unknown). It is intended for health-check dashboards
// and simple polling integrations.
//
// Endpoints:
//
//	GET /status        – returns the full job list as a JSON array
//	GET /status/{job}  – returns a single job by name, 404 if not found
//
// All responses use Content-Type: application/json. On error, the body
// contains a JSON object with a single "error" string field.
package api
