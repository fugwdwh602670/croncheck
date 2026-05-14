package probe

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestRecord_And_Get_Alive(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)

	if err := s.Record("backup", 5*time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Status != StatusAlive {
		t.Errorf("expected alive, got %s", e.Status)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	e, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected false for unknown job")
	}
	if e.Status != StatusUnknown {
		t.Errorf("expected unknown status, got %s", e.Status)
	}
}

func TestGet_DeadAfterTTL(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)
	_ = s.Record("sync", 1*time.Minute)

	s.now = fixedNow(base.Add(2 * time.Minute))
	e, ok := s.Get("sync")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Status != StatusDead {
		t.Errorf("expected dead, got %s", e.Status)
	}
}

func TestRecord_EmptyJobReturnsError(t *testing.T) {
	s := New()
	if err := s.Record("", time.Minute); err == nil {
		t.Fatal("expected error for empty job")
	}
}

func TestRecord_NonPositiveTTLReturnsError(t *testing.T) {
	s := New()
	if err := s.Record("job", 0); err == nil {
		t.Fatal("expected error for zero ttl")
	}
	if err := s.Record("job", -time.Second); err == nil {
		t.Fatal("expected error for negative ttl")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Record("a", time.Minute)
	_ = s.Record("b", time.Minute)
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}
