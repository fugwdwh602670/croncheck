package budget

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func makeStore(t *testing.T) *Store {
	t.Helper()
	return New()
}

func TestSet_And_Get(t *testing.T) {
	s := makeStore(t)
	if err := s.Set("backup", 5, time.Hour); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Limit != 5 || e.Window != time.Hour {
		t.Errorf("unexpected entry: %+v", e)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := makeStore(t)
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestSet_EmptyJobReturnsError(t *testing.T) {
	s := makeStore(t)
	if err := s.Set("", 3, time.Minute); err == nil {
		t.Fatal("expected error for empty job")
	}
}

func TestSet_NonPositiveLimitReturnsError(t *testing.T) {
	s := makeStore(t)
	if err := s.Set("backup", 0, time.Minute); err == nil {
		t.Fatal("expected error for zero limit")
	}
}

func TestSet_NonPositiveWindowReturnsError(t *testing.T) {
	s := makeStore(t)
	if err := s.Set("backup", 3, 0); err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestAllow_WithinBudget(t *testing.T) {
	s := makeStore(t)
	_ = s.Set("backup", 3, time.Hour)
	for i := 0; i < 3; i++ {
		if !s.Allow("backup") {
			t.Fatalf("expected Allow=true on call %d", i+1)
		}
	}
}

func TestAllow_ExceedsBudget(t *testing.T) {
	s := makeStore(t)
	_ = s.Set("backup", 2, time.Hour)
	s.Allow("backup")
	s.Allow("backup")
	if s.Allow("backup") {
		t.Fatal("expected Allow=false after budget exhausted")
	}
}

func TestAllow_ResetsAfterWindow(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	s := New()
	s.now = fixedNow(base)
	_ = s.Set("backup", 1, time.Hour)
	s.Allow("backup") // consume budget
	if s.Allow("backup") {
		t.Fatal("expected budget exhausted")
	}
	// advance past window
	s.now = fixedNow(base.Add(2 * time.Hour))
	if !s.Allow("backup") {
		t.Fatal("expected budget reset after window")
	}
}

func TestAllow_UnknownJobAlwaysAllowed(t *testing.T) {
	s := makeStore(t)
	if !s.Allow("ghost") {
		t.Fatal("expected unknown job to be allowed")
	}
}

func TestRemove_ClearsEntry(t *testing.T) {
	s := makeStore(t)
	_ = s.Set("backup", 5, time.Hour)
	s.Remove("backup")
	_, ok := s.Get("backup")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := makeStore(t)
	_ = s.Set("a", 1, time.Minute)
	_ = s.Set("b", 2, time.Minute)
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}
