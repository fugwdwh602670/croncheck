package owner

import (
	"testing"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	if err := s.Set("backup", "platform-team", "platform@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Owner != "platform-team" {
		t.Errorf("got owner %q, want %q", e.Owner, "platform-team")
	}
	if e.Contact != "platform@example.com" {
		t.Errorf("got contact %q, want %q", e.Contact, "platform@example.com")
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("expected false for unknown job")
	}
}

func TestSet_EmptyJobReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("", "team", ""); err == nil {
		t.Error("expected error for empty job name")
	}
}

func TestSet_EmptyOwnerReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("backup", "", ""); err == nil {
		t.Error("expected error for empty owner")
	}
}

func TestRemove_ClearsEntry(t *testing.T) {
	s := New()
	_ = s.Set("backup", "team", "")
	s.Remove("backup")
	_, ok := s.Get("backup")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Set("job-a", "team-a", "")
	_ = s.Set("job-b", "team-b", "b@example.com")
	all := s.All()
	if len(all) != 2 {
		t.Errorf("got %d entries, want 2", len(all))
	}
}

func TestSet_IsolatesMutation(t *testing.T) {
	s := New()
	_ = s.Set("job", "team-v1", "v1@example.com")
	_ = s.Set("job", "team-v2", "v2@example.com")
	e, _ := s.Get("job")
	if e.Owner != "team-v2" {
		t.Errorf("expected overwrite, got %q", e.Owner)
	}
}
