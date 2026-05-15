package expiry

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestSet_And_Get(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)

	if err := s.Set("backup", 10*time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Job != "backup" {
		t.Errorf("job = %q, want backup", e.Job)
	}
	if !e.Deadline.Equal(base.Add(10 * time.Minute)) {
		t.Errorf("deadline mismatch")
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
		t.Fatal("expected error for empty job")
	}
}

func TestSet_NonPositiveTTLReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("job", 0); err == nil {
		t.Fatal("expected error for zero ttl")
	}
	if err := s.Set("job", -time.Second); err == nil {
		t.Fatal("expected error for negative ttl")
	}
}

func TestIsExpired_NotYetExpired(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)
	_ = s.Set("job", 5*time.Minute)

	if s.IsExpired("job") {
		t.Fatal("expected job to not be expired yet")
	}
}

func TestIsExpired_AfterDeadline(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)
	_ = s.Set("job", 5*time.Minute)

	s.now = fixedNow(base.Add(6 * time.Minute))
	if !s.IsExpired("job") {
		t.Fatal("expected job to be expired")
	}
}

func TestIsExpired_UnknownJob(t *testing.T) {
	s := New()
	if s.IsExpired("ghost") {
		t.Fatal("unknown job should not be considered expired")
	}
}

func TestRemove_ClearsEntry(t *testing.T) {
	s := New()
	_ = s.Set("job", time.Minute)
	s.Remove("job")
	_, ok := s.Get("job")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Set("a", time.Minute)
	_ = s.Set("b", 2*time.Minute)
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}
