package tags

import (
	"sort"
	"testing"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	s.Set("backup", map[string]string{"env": "prod", "team": "ops"})

	got, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected tags to exist")
	}
	if got["env"] != "prod" || got["team"] != "ops" {
		t.Errorf("unexpected tags: %v", got)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("unknown")
	if ok {
		t.Fatal("expected ok=false for unknown job")
	}
}

func TestSet_IsolatesMutation(t *testing.T) {
	s := New()
	orig := map[string]string{"env": "prod"}
	s.Set("job", orig)
	orig["env"] = "staging" // mutate original

	got, _ := s.Get("job")
	if got["env"] != "prod" {
		t.Errorf("store should not reflect external mutation")
	}
}

func TestRemove_ClearsTags(t *testing.T) {
	s := New()
	s.Set("job", map[string]string{"env": "prod"})
	s.Remove("job")
	_, ok := s.Get("job")
	if ok {
		t.Fatal("expected tags to be removed")
	}
}

func TestFilter_MatchesSubset(t *testing.T) {
	s := New()
	s.Set("job-a", map[string]string{"env": "prod", "team": "ops"})
	s.Set("job-b", map[string]string{"env": "prod", "team": "dev"})
	s.Set("job-c", map[string]string{"env": "staging", "team": "ops"})

	result := s.Filter(map[string]string{"env": "prod"})
	sort.Strings(result)
	if len(result) != 2 || result[0] != "job-a" || result[1] != "job-b" {
		t.Errorf("unexpected filter result: %v", result)
	}
}

func TestFilter_NoMatch(t *testing.T) {
	s := New()
	s.Set("job-a", map[string]string{"env": "prod"})
	result := s.Filter(map[string]string{"env": "staging"})
	if len(result) != 0 {
		t.Errorf("expected no results, got %v", result)
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	s.Set("job-a", map[string]string{"env": "prod"})
	s.Set("job-b", map[string]string{"env": "staging"})
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestValidate_EmptyKey(t *testing.T) {
	err := Validate(map[string]string{"": "value"})
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestValidate_EmptyValue(t *testing.T) {
	err := Validate(map[string]string{"env": ""})
	if err == nil {
		t.Fatal("expected error for empty value")
	}
}

func TestValidate_Valid(t *testing.T) {
	err := Validate(map[string]string{"env": "prod", "team": "ops"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
