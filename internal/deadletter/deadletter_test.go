package deadletter

import (
	"testing"
	"time"
)

func TestAdd_StoresEntry(t *testing.T) {
	s := New(10)
	s.Add("backup", "connection refused", `{"job":"backup"}`, 1)

	all := s.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(all))
	}
	e := all[0]
	if e.Job != "backup" {
		t.Errorf("job: got %q, want %q", e.Job, "backup")
	}
	if e.Attempts != 1 {
		t.Errorf("attempts: got %d, want 1", e.Attempts)
	}
	if e.FailedAt.IsZero() {
		t.Error("FailedAt should not be zero")
	}
}

func TestAdd_RespectsLimit(t *testing.T) {
	s := New(3)
	for i := 0; i < 5; i++ {
		s.Add("job", "err", "", i)
	}

	if s.Count() != 3 {
		t.Errorf("expected 3 entries, got %d", s.Count())
	}
	// Oldest should have been evicted; newest attempt index is 4.
	all := s.All()
	if all[len(all)-1].Attempts != 4 {
		t.Errorf("expected last attempts=4, got %d", all[len(all)-1].Attempts)
	}
}

func TestAdd_DefaultLimit(t *testing.T) {
	s := New(0)
	if s.limit != 100 {
		t.Errorf("expected default limit 100, got %d", s.limit)
	}
}

func TestRemove_ClearsJob(t *testing.T) {
	s := New(10)
	s.Add("alpha", "err", "", 1)
	s.Add("beta", "err", "", 1)
	s.Add("alpha", "err", "", 2)

	s.Remove("alpha")

	all := s.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 entry after remove, got %d", len(all))
	}
	if all[0].Job != "beta" {
		t.Errorf("expected remaining job=beta, got %q", all[0].Job)
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New(10)
	s.Add("job", "err", "", 1)

	snap := s.All()
	snap[0].Job = "mutated"

	original := s.All()
	if original[0].Job == "mutated" {
		t.Error("All() should return an independent snapshot")
	}
}

func TestAdd_TimestampIsRecent(t *testing.T) {
	before := time.Now().UTC()
	s := New(10)
	s.Add("job", "err", "", 1)
	after := time.Now().UTC()

	e := s.All()[0]
	if e.FailedAt.Before(before) || e.FailedAt.After(after) {
		t.Errorf("FailedAt %v not in expected range [%v, %v]", e.FailedAt, before, after)
	}
}
