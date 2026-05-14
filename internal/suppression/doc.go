// Package suppression manages per-job alert suppression thresholds.
//
// A suppression rule specifies the minimum number of consecutive missed
// runs that must occur before an alert is forwarded. This prevents noisy
// notifications for jobs that occasionally miss a single run.
//
// Usage:
//
//	s := suppression.New()
//	_ = s.Set("nightly-backup", 3)
//	if !s.IsSuppressed("nightly-backup", consecMisses) {
//	    // send alert
//	}
package suppression
