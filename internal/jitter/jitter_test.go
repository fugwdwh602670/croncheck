package jitter

import (
	"testing"
	"time"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	if err := s.Set("backup", 5*time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if d != 5*time.Minute {
		t.Fatalf("expected 5m, got %v", d)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("nonexistent")
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

func TestSet_NonPositiveDurationReturnsError(t *testing.T) {
	s := New()
	for _, d := range []time.Duration{0, -time.Second} {
		if err := s.Set("job", d); err == nil {
			t.Fatalf("expected error for duration %v", d)
		}
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
	all["c"] = 3 * time.Minute
	if _, ok := s.Get("c"); ok {
		t.Fatal("mutation of snapshot should not affect store")
	}
}

func TestGrace_FallsBackToDefault(t *testing.T) {
	s := New()
	defaultGrace := 30 * time.Second
	if g := s.Grace("unknown", defaultGrace); g != defaultGrace {
		t.Fatalf("expected default grace %v, got %v", defaultGrace, g)
	}
}

func TestGrace_UsesStoredValue(t *testing.T) {
	s := New()
	_ = s.Set("myjob", 10*time.Minute)
	if g := s.Grace("myjob", 30*time.Second); g != 10*time.Minute {
		t.Fatalf("expected stored grace 10m, got %v", g)
	}
}
