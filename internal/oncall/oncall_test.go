package oncall

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestSet_And_Get(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)

	if err := s.Set("backup", "alice@example.com", time.Hour); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to be present")
	}
	if e.Contact != "alice@example.com" {
		t.Errorf("contact = %q, want %q", e.Contact, "alice@example.com")
	}
}

func TestGet_ExpiredEntry(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)
	_ = s.Set("backup", "alice@example.com", time.Millisecond)

	s.now = fixedNow(base.Add(time.Second))
	_, ok := s.Get("backup")
	if ok {
		t.Fatal("expected expired entry to be absent")
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("nope")
	if ok {
		t.Fatal("expected false for unknown job")
	}
}

func TestSet_ValidationErrors(t *testing.T) {
	s := New()
	if err := s.Set("", "c", time.Hour); err == nil {
		t.Error("expected error for empty job")
	}
	if err := s.Set("job", "", time.Hour); err == nil {
		t.Error("expected error for empty contact")
	}
	if err := s.Set("job", "c", -time.Second); err == nil {
		t.Error("expected error for non-positive duration")
	}
}

func TestRemove_ClearsEntry(t *testing.T) {
	s := New()
	_ = s.Set("backup", "alice@example.com", time.Hour)
	s.Remove("backup")
	_, ok := s.Get("backup")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestAll_ReturnsOnlyActive(t *testing.T) {
	s := New()
	base := time.Now()
	s.now = fixedNow(base)
	_ = s.Set("job1", "alice@example.com", time.Hour)
	_ = s.Set("job2", "bob@example.com", time.Millisecond)

	s.now = fixedNow(base.Add(time.Second))
	all := s.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 active entry, got %d", len(all))
	}
	if all[0].Job != "job1" {
		t.Errorf("expected job1, got %q", all[0].Job)
	}
}
