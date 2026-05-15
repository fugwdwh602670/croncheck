package webhook

import (
	"testing"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	if err := s.Set("backup", "https://example.com/hook", "topsecret"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.URL != "https://example.com/hook" {
		t.Errorf("url = %q, want %q", e.URL, "https://example.com/hook")
	}
	if e.Secret != "topsecret" {
		t.Errorf("secret = %q, want %q", e.Secret, "topsecret")
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
	if err := s.Set("", "https://example.com", ""); err == nil {
		t.Fatal("expected error for empty job")
	}
}

func TestSet_EmptyURLReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("backup", "", ""); err == nil {
		t.Fatal("expected error for empty url")
	}
}

func TestRemove_ClearsEntry(t *testing.T) {
	s := New()
	_ = s.Set("backup", "https://example.com/hook", "")
	s.Remove("backup")
	_, ok := s.Get("backup")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Set("job-a", "https://a.example.com", "")
	_ = s.Set("job-b", "https://b.example.com", "s")
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestSet_Overwrites(t *testing.T) {
	s := New()
	_ = s.Set("job-a", "https://old.example.com", "")
	_ = s.Set("job-a", "https://new.example.com", "newsecret")
	e, _ := s.Get("job-a")
	if e.URL != "https://new.example.com" {
		t.Errorf("url = %q, want new url", e.URL)
	}
}
