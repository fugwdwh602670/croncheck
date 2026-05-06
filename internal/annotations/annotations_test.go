package annotations

import (
	"testing"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	s.Set("backup", map[string]string{"owner": "ops", "env": "prod"})
	a, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected annotations to exist")
	}
	if a["owner"] != "ops" || a["env"] != "prod" {
		t.Fatalf("unexpected annotations: %v", a)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected false for unknown job")
	}
}

func TestSet_IsolatesMutation(t *testing.T) {
	s := New()
	orig := map[string]string{"k": "v"}
	s.Set("job", orig)
	orig["k"] = "mutated"
	a, _ := s.Get("job")
	if a["k"] != "v" {
		t.Fatal("store should not be affected by external mutation")
	}
}

func TestSet_EmptyMapClearsAnnotations(t *testing.T) {
	s := New()
	s.Set("job", map[string]string{"x": "y"})
	s.Set("job", nil)
	_, ok := s.Get("job")
	if ok {
		t.Fatal("expected annotations to be cleared")
	}
}

func TestRemove_ClearsAnnotations(t *testing.T) {
	s := New()
	s.Set("job", map[string]string{"a": "b"})
	s.Remove("job")
	_, ok := s.Get("job")
	if ok {
		t.Fatal("expected annotations to be removed")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	s.Set("j1", map[string]string{"k": "1"})
	s.Set("j2", map[string]string{"k": "2"})
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// Mutating the snapshot must not affect the store.
	all["j1"]["k"] = "mutated"
	a, _ := s.Get("j1")
	if a["k"] != "1" {
		t.Fatal("store affected by snapshot mutation")
	}
}
