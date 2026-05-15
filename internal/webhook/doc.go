// Package webhook provides per-job webhook registration for croncheck.
//
// Each cron job may optionally have a dedicated webhook URL. When an alert
// fires for that job, the webhook URL is POSTed a JSON payload containing
// the job name, alert reason, and timestamp. An optional shared secret is
// stored alongside the URL so callers can verify HMAC signatures on the
// receiving end.
//
// The HTTP handler exposes a simple REST interface:
//
//	GET    /webhooks          – list all registered webhooks
//	GET    /webhooks?job=name – retrieve the webhook for a specific job
//	PUT    /webhooks          – register or update a webhook (JSON body)
//	DELETE /webhooks?job=name – remove a webhook registration
package webhook
