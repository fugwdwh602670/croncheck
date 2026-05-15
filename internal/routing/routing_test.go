package routing

import (
	"testing"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	if err := s.Set("backup", "slack"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ch, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected rule to exist")
	}
	if ch != "slack" {
		t.Fatalf("expected slack, got %s", ch)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("unknown")
	if ok {
		t.Fatal("expected no rule for unknown job")
	}
}

func TestSet_EmptyJobReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("", "slack"); err == nil {
		t.Fatal("expected error for empty job")
	}
}

func TestSet_EmptyChannelReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("backup", ""); err == nil {
		t.Fatal("expected error for empty channel")
	}
}

func TestSet_Overwrites(t *testing.T) {
	s := New()
	_ = s.Set("backup", "slack")
	_ = s.Set("backup", "pagerduty")
	ch, _ := s.Get("backup")
	if ch != "pagerduty" {
		t.Fatalf("expected pagerduty, got %s", ch)
	}
}

func TestRemove_ClearsRule(t *testing.T) {
	s := New()
	_ = s.Set("backup", "slack")
	s.Remove("backup")
	_, ok := s.Get("backup")
	if ok {
		t.Fatal("expected rule to be removed")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Set("backup", "slack")
	_ = s.Set("deploy", "email")
	rules := s.All()
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
}
