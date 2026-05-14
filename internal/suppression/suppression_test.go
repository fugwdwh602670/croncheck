package suppression

import (
	"testing"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	if err := s.Set("backup", 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected rule to exist")
	}
	if r.MinConsecMisses != 3 {
		t.Errorf("expected 3, got %d", r.MinConsecMisses)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected no rule for unknown job")
	}
}

func TestSet_EmptyJobReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("", 2); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestSet_NonPositiveThresholdReturnsError(t *testing.T) {
	s := New()
	for _, v := range []int{0, -1} {
		if err := s.Set("job", v); err == nil {
			t.Errorf("expected error for min_consec_misses=%d", v)
		}
	}
}

func TestIsSuppressed_BelowThreshold(t *testing.T) {
	s := New()
	_ = s.Set("backup", 3)
	if !s.IsSuppressed("backup", 2) {
		t.Error("expected alert to be suppressed when misses < threshold")
	}
}

func TestIsSuppressed_AtThreshold(t *testing.T) {
	s := New()
	_ = s.Set("backup", 3)
	if s.IsSuppressed("backup", 3) {
		t.Error("expected alert NOT suppressed when misses == threshold")
	}
}

func TestIsSuppressed_NoRule(t *testing.T) {
	s := New()
	if s.IsSuppressed("nojob", 10) {
		t.Error("expected not suppressed when no rule exists")
	}
}

func TestRemove_ClearsRule(t *testing.T) {
	s := New()
	_ = s.Set("backup", 2)
	s.Remove("backup")
	_, ok := s.Get("backup")
	if ok {
		t.Fatal("expected rule to be removed")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Set("a", 1)
	_ = s.Set("b", 5)
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 rules, got %d", len(all))
	}
}
