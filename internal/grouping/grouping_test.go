package grouping

import (
	"testing"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	if err := s.Set("backup", "ops"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	g, ok := s.Get("backup")
	if !ok || g != "ops" {
		t.Fatalf("expected ops, got %q ok=%v", g, ok)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("unknown")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestSet_EmptyJobReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("", "ops"); err == nil {
		t.Fatal("expected error for empty job")
	}
}

func TestSet_EmptyGroupReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("backup", ""); err == nil {
		t.Fatal("expected error for empty group")
	}
}

func TestRemove_ClearsGroup(t *testing.T) {
	s := New()
	_ = s.Set("backup", "ops")
	s.Remove("backup")
	_, ok := s.Get("backup")
	if ok {
		t.Fatal("expected job to be removed")
	}
}

func TestJobsInGroup_ReturnsMembers(t *testing.T) {
	s := New()
	_ = s.Set("backup", "ops")
	_ = s.Set("deploy", "ops")
	_ = s.Set("report", "finance")
	jobs := s.JobsInGroup("ops")
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestJobsInGroup_EmptyForUnknownGroup(t *testing.T) {
	s := New()
	if jobs := s.JobsInGroup("unknown"); len(jobs) != 0 {
		t.Fatalf("expected empty, got %v", jobs)
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Set("backup", "ops")
	_ = s.Set("report", "finance")
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	all["mutate"] = "should-not-affect-store"
	if _, ok := s.Get("mutate"); ok {
		t.Fatal("snapshot mutation leaked into store")
	}
}
