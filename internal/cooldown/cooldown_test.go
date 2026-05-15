package cooldown

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestSet_And_Get(t *testing.T) {
	s := New()
	if err := s.Set("backup", 5*time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Duration != 5*time.Minute {
		t.Fatalf("want 5m, got %v", e.Duration)
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

func TestSet_NonPositiveDurationReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("backup", 0); err == nil {
		t.Fatal("expected error for zero duration")
	}
	if err := s.Set("backup", -time.Second); err == nil {
		t.Fatal("expected error for negative duration")
	}
}

func TestAllow_NoCooldownConfigured(t *testing.T) {
	s := New()
	if err := s.Allow("unconfigured"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_FirstHeartbeat(t *testing.T) {
	s := New()
	_ = s.Set("backup", time.Minute)
	if err := s.Allow("backup"); err != nil {
		t.Fatalf("first heartbeat should be allowed, got %v", err)
	}
}

func TestAllow_WithinCooldown(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	s := New()
	s.now = fixedNow(base)
	_ = s.Set("backup", time.Minute)
	_ = s.Allow("backup") // first – sets LastSeen

	// advance by 30s – still within cooldown
	s.now = fixedNow(base.Add(30 * time.Second))
	if err := s.Allow("backup"); err != ErrTooSoon {
		t.Fatalf("want ErrTooSoon, got %v", err)
	}
}

func TestAllow_AfterCooldown(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	s := New()
	s.now = fixedNow(base)
	_ = s.Set("backup", time.Minute)
	_ = s.Allow("backup")

	s.now = fixedNow(base.Add(90 * time.Second))
	if err := s.Allow("backup"); err != nil {
		t.Fatalf("expected nil after cooldown, got %v", err)
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

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Set("a", time.Minute)
	_ = s.Set("b", 2*time.Minute)
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("want 2 entries, got %d", len(all))
	}
}
