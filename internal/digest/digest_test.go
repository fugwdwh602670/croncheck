package digest

import (
	"testing"
	"time"
)

func makeStates() []JobState {
	return []JobState{
		{Name: "alpha", MissedCount: 0, Healthy: true, LastSeen: time.Now().Add(-1 * time.Minute)},
		{Name: "beta", MissedCount: 3, Healthy: false, LastSeen: time.Now().Add(-10 * time.Minute)},
	}
}

func TestBuild_StoresEntries(t *testing.T) {
	s := New()
	s.Build(makeStates())
	entries := s.All()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestBuild_FieldsPopulated(t *testing.T) {
	s := New()
	s.Build(makeStates())
	entries := s.All()
	var beta Entry
	for _, e := range entries {
		if e.Job == "beta" {
			beta = e
		}
	}
	if beta.Job == "" {
		t.Fatal("beta entry not found")
	}
	if beta.MissedCount != 3 {
		t.Errorf("expected missed_count=3, got %d", beta.MissedCount)
	}
	if beta.Healthy {
		t.Error("expected healthy=false")
	}
	if beta.CapturedAt.IsZero() {
		t.Error("expected CapturedAt to be set")
	}
}

func TestBuild_ReplacesExistingEntries(t *testing.T) {
	s := New()
	s.Build(makeStates())
	s.Build([]JobState{{Name: "gamma", Healthy: true}})
	entries := s.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after rebuild, got %d", len(entries))
	}
	if entries[0].Job != "gamma" {
		t.Errorf("expected gamma, got %s", entries[0].Job)
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	s.Build(makeStates())
	a := s.All()
	a[0].Job = "mutated"
	b := s.All()
	if b[0].Job == "mutated" {
		t.Error("All() should return an independent copy")
	}
}

func TestBuiltAt_ZeroBeforeBuild(t *testing.T) {
	s := New()
	if !s.BuiltAt().IsZero() {
		t.Error("expected zero BuiltAt before first Build")
	}
}

func TestBuiltAt_SetAfterBuild(t *testing.T) {
	s := New()
	before := time.Now().UTC()
	s.Build(makeStates())
	if s.BuiltAt().Before(before) {
		t.Error("BuiltAt should be >= time of Build call")
	}
}
