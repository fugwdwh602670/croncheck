package pausejob

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestIsPaused_ActivePause(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)

	if err := s.Pause("backup", "maintenance", time.Hour); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.IsPaused("backup") {
		t.Fatal("expected job to be paused")
	}
}

func TestIsPaused_ExpiredPause(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)
	_ = s.Pause("backup", "maintenance", time.Minute)

	s.now = fixedNow(base.Add(2 * time.Minute))
	if s.IsPaused("backup") {
		t.Fatal("expected pause to have expired")
	}
}

func TestIsPaused_UnknownJob(t *testing.T) {
	s := New()
	if s.IsPaused("unknown") {
		t.Fatal("expected false for unknown job")
	}
}

func TestResume_ClearsPause(t *testing.T) {
	s := New()
	s.now = fixedNow(time.Now())
	_ = s.Pause("sync", "", time.Hour)
	s.Resume("sync")
	if s.IsPaused("sync") {
		t.Fatal("expected job to no longer be paused after resume")
	}
}

func TestAll_ReturnsOnlyActive(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)
	_ = s.Pause("job-a", "r", time.Hour)
	_ = s.Pause("job-b", "r", time.Minute)

	s.now = fixedNow(base.Add(90 * time.Second))
	all := s.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 active pause, got %d", len(all))
	}
	if all[0].Job != "job-a" {
		t.Fatalf("expected job-a, got %s", all[0].Job)
	}
}

func TestPause_EmptyJobReturnsError(t *testing.T) {
	s := New()
	if err := s.Pause("", "r", time.Hour); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestPause_NegativeDurationReturnsError(t *testing.T) {
	s := New()
	if err := s.Pause("job", "r", -time.Second); err == nil {
		t.Fatal("expected error for non-positive duration")
	}
}
