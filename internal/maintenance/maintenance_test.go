package maintenance

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestSet_And_IsActive(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)

	if err := s.Set("backup", time.Hour); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.IsActive("backup") {
		t.Error("expected window to be active")
	}
}

func TestIsActive_ExpiredWindow(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)
	_ = s.Set("backup", time.Minute)

	// advance time past window
	s.now = fixedNow(base.Add(2 * time.Minute))
	if s.IsActive("backup") {
		t.Error("expected window to be expired")
	}
}

func TestIsActive_UnknownJob(t *testing.T) {
	s := New()
	if s.IsActive("unknown") {
		t.Error("expected false for unknown job")
	}
}

func TestSet_EmptyJobReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("", time.Hour); err == nil {
		t.Error("expected error for empty job name")
	}
}

func TestSet_NegativeDurationReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("backup", -time.Minute); err == nil {
		t.Error("expected error for negative duration")
	}
}

func TestRemove_ClearsWindow(t *testing.T) {
	s := New()
	s.now = fixedNow(time.Now())
	_ = s.Set("backup", time.Hour)
	s.Remove("backup")
	if s.IsActive("backup") {
		t.Error("expected window to be cleared after remove")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	s.now = fixedNow(time.Now())
	_ = s.Set("job-a", time.Hour)
	_ = s.Set("job-b", 30*time.Minute)

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(all))
	}
}

func TestAll_EmptyStore(t *testing.T) {
	s := New()
	if got := s.All(); len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}
