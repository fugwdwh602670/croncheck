package scheduler

import (
	"testing"
	"time"

	"croncheck/internal/config"
)

func makeScheduler(jobs ...config.Job) *Scheduler {
	return New(jobs)
}

func TestHeartbeat_KnownJob(t *testing.T) {
	s := makeScheduler(config.Job{Name: "backup", Schedule: "0 2 * * *"})
	if !s.Heartbeat("backup") {
		t.Fatal("expected heartbeat to return true for known job")
	}
	s.mu.RLock()
	job := s.jobs["backup"]
	s.mu.RUnlock()
	if job.LastSeen.IsZero() {
		t.Error("expected LastSeen to be set after heartbeat")
	}
	if job.Missed {
		t.Error("expected Missed to be false after heartbeat")
	}
}

func TestHeartbeat_UnknownJob(t *testing.T) {
	s := makeScheduler()
	if s.Heartbeat("ghost") {
		t.Fatal("expected heartbeat to return false for unknown job")
	}
}

func TestCheckMissed_NoHeartbeat(t *testing.T) {
	// Jobs with zero LastSeen should not be flagged
	s := makeScheduler(config.Job{Name: "report", Schedule: "*/5 * * * *"})
	missed := s.CheckMissed(time.Now())
	if len(missed) != 0 {
		t.Errorf("expected 0 missed jobs, got %d", len(missed))
	}
}

func TestCheckMissed_Overdue(t *testing.T) {
	s := makeScheduler(config.Job{Name: "sync", Schedule: "*/5 * * * *"})

	// Simulate a heartbeat 20 minutes ago — well past the 5-min schedule + 2-min grace
	s.jobs["sync"].LastSeen = time.Now().Add(-20 * time.Minute)

	missed := s.CheckMissed(time.Now())
	if len(missed) != 1 {
		t.Fatalf("expected 1 missed job, got %d", len(missed))
	}
	if missed[0].Name != "sync" {
		t.Errorf("expected missed job 'sync', got '%s'", missed[0].Name)
	}
}

func TestCheckMissed_WithinGrace(t *testing.T) {
	s := makeScheduler(config.Job{Name: "ping", Schedule: "*/5 * * * *"})

	// Heartbeat just 6 minutes ago — next run was 1 min ago, still within 2-min grace
	s.jobs["ping"].LastSeen = time.Now().Add(-6 * time.Minute)

	missed := s.CheckMissed(time.Now())
	if len(missed) != 0 {
		t.Errorf("expected 0 missed jobs within grace, got %d", len(missed))
	}
}
