package notify_policy

import (
	"testing"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	p := Policy{MinSeverity: SeverityWarning, MinConsecMisses: 2}
	if err := s.Set("backup", p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected policy to exist")
	}
	if got != p {
		t.Errorf("got %+v, want %+v", got, p)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected no policy for unknown job")
	}
}

func TestSet_EmptyJobReturnsError(t *testing.T) {
	s := New()
	err := s.Set("", Policy{MinSeverity: SeverityInfo, MinConsecMisses: 1})
	if err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestSet_InvalidSeverityReturnsError(t *testing.T) {
	s := New()
	err := s.Set("myjob", Policy{MinSeverity: "extreme", MinConsecMisses: 1})
	if err == nil {
		t.Fatal("expected error for invalid severity")
	}
}

func TestSet_ZeroConsecMissesReturnsError(t *testing.T) {
	s := New()
	err := s.Set("myjob", Policy{MinSeverity: SeverityCritical, MinConsecMisses: 0})
	if err == nil {
		t.Fatal("expected error for zero min_consec_misses")
	}
}

func TestRemove_ClearsPolicy(t *testing.T) {
	s := New()
	_ = s.Set("cleanup", Policy{MinSeverity: SeverityInfo, MinConsecMisses: 1})
	s.Remove("cleanup")
	_, ok := s.Get("cleanup")
	if ok {
		t.Fatal("expected policy to be removed")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Set("jobA", Policy{MinSeverity: SeverityInfo, MinConsecMisses: 1})
	_ = s.Set("jobB", Policy{MinSeverity: SeverityCritical, MinConsecMisses: 3})
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	all["jobA"] = Policy{MinSeverity: SeverityCritical, MinConsecMisses: 99}
	got, _ := s.Get("jobA")
	if got.MinConsecMisses == 99 {
		t.Error("All() should return an isolated copy")
	}
}
