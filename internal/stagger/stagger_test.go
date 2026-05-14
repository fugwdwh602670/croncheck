package stagger

import (
	"testing"
	"time"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	if err := s.Set("backup", 5*time.Second); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if d != 5*time.Second {
		t.Fatalf("expected 5s, got %v", d)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestSet_EmptyJobReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("", time.Second); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestSet_NonPositiveDurationReturnsError(t *testing.T) {
	s := New()
	for _, d := range []time.Duration{0, -time.Second} {
		if err := s.Set("job", d); err == nil {
			t.Fatalf("expected error for delay %v", d)
		}
	}
}

func TestRemove_ClearsEntry(t *testing.T) {
	s := New()
	_ = s.Set("cleanup", time.Minute)
	s.Remove("cleanup")
	_, ok := s.Get("cleanup")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Set("a", time.Second)
	_ = s.Set("b", 2*time.Second)
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestAll_IsolatesMutation(t *testing.T) {
	s := New()
	_ = s.Set("x", time.Second)
	all := s.All()
	all[0].Delay = 99 * time.Hour
	d, _ := s.Get("x")
	if d != time.Second {
		t.Fatal("snapshot mutation affected internal state")
	}
}
