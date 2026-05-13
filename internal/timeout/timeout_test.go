package timeout

import (
	"testing"
	"time"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	if err := s.Set("backup", 5*time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Threshold != 5*time.Minute {
		t.Errorf("got threshold %s, want 5m", e.Threshold)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestSet_EmptyJobReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("", time.Minute); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestSet_NegativeThresholdReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("backup", -time.Second); err == nil {
		t.Fatal("expected error for non-positive threshold")
	}
}

func TestSet_ZeroThresholdReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("backup", 0); err == nil {
		t.Fatal("expected error for zero threshold")
	}
}

func TestRemove_ClearsEntry(t *testing.T) {
	s := New()
	_ = s.Set("backup", time.Minute)
	s.Remove("backup")
	_, ok := s.Get("backup")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestIsExceeded_NotExceeded(t *testing.T) {
	s := New()
	_ = s.Set("backup", 10*time.Minute)
	now := time.Now()
	lastSeen := now.Add(-5 * time.Minute)
	if s.IsExceeded("backup", lastSeen, now) {
		t.Error("expected threshold not to be exceeded")
	}
}

func TestIsExceeded_Exceeded(t *testing.T) {
	s := New()
	_ = s.Set("backup", 10*time.Minute)
	now := time.Now()
	lastSeen := now.Add(-15 * time.Minute)
	if !s.IsExceeded("backup", lastSeen, now) {
		t.Error("expected threshold to be exceeded")
	}
}

func TestIsExceeded_NoThreshold(t *testing.T) {
	s := New()
	if s.IsExceeded("ghost", time.Now().Add(-time.Hour), time.Now()) {
		t.Error("expected false when no threshold registered")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Set("job-a", time.Minute)
	_ = s.Set("job-b", 2*time.Minute)
	all := s.All()
	if len(all) != 2 {
		t.Errorf("got %d entries, want 2", len(all))
	}
}
