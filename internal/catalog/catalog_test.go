package catalog

import (
	"testing"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	e := Entry{Job: "backup", Description: "Daily backup", Owner: "ops", Schedule: "0 2 * * *"}
	if err := s.Set(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if got != e {
		t.Errorf("got %+v, want %+v", got, e)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("missing")
	if ok {
		t.Fatal("expected entry to be absent")
	}
}

func TestSet_EmptyJobReturnsError(t *testing.T) {
	s := New()
	if err := s.Set(Entry{}); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestRemove_ClearsEntry(t *testing.T) {
	s := New()
	_ = s.Set(Entry{Job: "nightly", Owner: "team-a"})
	s.Remove("nightly")
	_, ok := s.Get("nightly")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Set(Entry{Job: "job-a"})
	_ = s.Set(Entry{Job: "job-b"})
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestAll_EmptyStore(t *testing.T) {
	s := New()
	if got := s.All(); len(got) != 0 {
		t.Fatalf("expected empty slice, got %d entries", len(got))
	}
}

func TestSet_Overwrites(t *testing.T) {
	s := New()
	_ = s.Set(Entry{Job: "deploy", Owner: "old-owner"})
	_ = s.Set(Entry{Job: "deploy", Owner: "new-owner"})
	got, _ := s.Get("deploy")
	if got.Owner != "new-owner" {
		t.Errorf("expected new-owner, got %s", got.Owner)
	}
}
