package replay

import (
	"testing"
	"time"
)

var fixedNow = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

func makeStore() *Store {
	s := New()
	s.now = func() time.Time { return fixedNow }
	return s
}

func TestRequest_StoresEntry(t *testing.T) {
	s := makeStore()
	if err := s.Request("backup"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Job != "backup" {
		t.Errorf("expected job=backup, got %s", e.Job)
	}
	if e.Acked {
		t.Error("expected acked=false")
	}
	if !e.RequestedAt.Equal(fixedNow) {
		t.Errorf("unexpected requested_at: %v", e.RequestedAt)
	}
}

func TestRequest_EmptyJobReturnsError(t *testing.T) {
	s := makeStore()
	if err := s.Request(""); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestAck_MarksEntry(t *testing.T) {
	s := makeStore()
	_ = s.Request("sync")
	if err := s.Ack("sync"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, _ := s.Get("sync")
	if !e.Acked {
		t.Error("expected acked=true")
	}
	if !e.AckedAt.Equal(fixedNow) {
		t.Errorf("unexpected acked_at: %v", e.AckedAt)
	}
}

func TestAck_UnknownJobReturnsError(t *testing.T) {
	s := makeStore()
	if err := s.Ack("ghost"); err == nil {
		t.Fatal("expected error for unknown job")
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := makeStore()
	_, ok := s.Get("nope")
	if ok {
		t.Error("expected ok=false for unknown job")
	}
}

func TestRemove_ClearsEntry(t *testing.T) {
	s := makeStore()
	_ = s.Request("cleanup")
	s.Remove("cleanup")
	_, ok := s.Get("cleanup")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := makeStore()
	_ = s.Request("job-a")
	_ = s.Request("job-b")
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestAll_EmptyStore(t *testing.T) {
	s := makeStore()
	if got := s.All(); len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}
