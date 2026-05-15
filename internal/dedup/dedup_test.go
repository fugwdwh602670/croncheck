package dedup

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestIsDuplicate_FirstCall_ReturnsFalse(t *testing.T) {
	s := New()
	if s.IsDuplicate("backup", time.Minute) {
		t.Fatal("expected false on first call")
	}
}

func TestIsDuplicate_SecondCall_ReturnsTrue(t *testing.T) {
	s := New()
	s.IsDuplicate("backup", time.Minute)
	if !s.IsDuplicate("backup", time.Minute) {
		t.Fatal("expected true on second call within window")
	}
}

func TestIsDuplicate_AfterWindowExpiry_ReturnsFalse(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	s := New()
	s.now = fixedNow(base)
	s.IsDuplicate("backup", time.Minute)

	// advance past the window
	s.now = fixedNow(base.Add(2 * time.Minute))
	if s.IsDuplicate("backup", time.Minute) {
		t.Fatal("expected false after window expiry")
	}
}

func TestReset_AllowsImmediately(t *testing.T) {
	s := New()
	s.IsDuplicate("backup", time.Minute)
	s.Reset("backup")
	if s.IsDuplicate("backup", time.Minute) {
		t.Fatal("expected false after reset")
	}
}

func TestCount_TracksSuppressions(t *testing.T) {
	s := New()
	s.IsDuplicate("backup", time.Minute) // original
	s.IsDuplicate("backup", time.Minute) // dup 1
	s.IsDuplicate("backup", time.Minute) // dup 2
	if got := s.Count("backup"); got != 3 {
		t.Fatalf("expected count 3, got %d", got)
	}
}

func TestCount_UnknownJob_ReturnsZero(t *testing.T) {
	s := New()
	if got := s.Count("unknown"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestAll_ReturnsActiveEntries(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	s := New()
	s.now = fixedNow(base)
	s.IsDuplicate("jobA", time.Minute)
	s.IsDuplicate("jobB", time.Minute)

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestAll_ExcludesExpired(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	s := New()
	s.now = fixedNow(base)
	s.IsDuplicate("jobA", time.Minute)

	s.now = fixedNow(base.Add(2 * time.Minute))
	all := s.All()
	if len(all) != 0 {
		t.Fatalf("expected 0 active entries after expiry, got %d", len(all))
	}
}
