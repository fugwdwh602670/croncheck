package dependency

import (
	"testing"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	if err := s.Set("job-b", []string{"job-a"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ups := s.Get("job-b")
	if len(ups) != 1 || ups[0] != "job-a" {
		t.Fatalf("expected [job-a], got %v", ups)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	if ups := s.Get("ghost"); ups != nil {
		t.Fatalf("expected nil, got %v", ups)
	}
}

func TestSet_SelfDependencyReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("job-a", []string{"job-a"}); err == nil {
		t.Fatal("expected error for self-dependency, got nil")
	}
}

func TestSet_IsolatesMutation(t *testing.T) {
	s := New()
	original := []string{"job-a", "job-b"}
	_ = s.Set("job-c", original)
	original[0] = "mutated"
	ups := s.Get("job-c")
	if ups[0] != "job-a" {
		t.Fatalf("store was mutated via original slice")
	}
}

func TestRemove_ClearsDependencies(t *testing.T) {
	s := New()
	_ = s.Set("job-b", []string{"job-a"})
	s.Remove("job-b")
	if ups := s.Get("job-b"); ups != nil {
		t.Fatalf("expected nil after remove, got %v", ups)
	}
}

func TestBlockedBy_AllHealthy(t *testing.T) {
	s := New()
	_ = s.Set("job-b", []string{"job-a"})
	blocked := s.BlockedBy("job-b", func(string) bool { return true })
	if len(blocked) != 0 {
		t.Fatalf("expected no blocked upstreams, got %v", blocked)
	}
}

func TestBlockedBy_UnhealthyUpstream(t *testing.T) {
	s := New()
	_ = s.Set("job-b", []string{"job-a", "job-c"})
	unhealthy := map[string]bool{"job-a": true}
	blocked := s.BlockedBy("job-b", func(j string) bool { return !unhealthy[j] })
	if len(blocked) != 1 || blocked[0] != "job-a" {
		t.Fatalf("expected [job-a], got %v", blocked)
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Set("job-b", []string{"job-a"})
	_ = s.Set("job-c", []string{"job-a", "job-b"})
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}
