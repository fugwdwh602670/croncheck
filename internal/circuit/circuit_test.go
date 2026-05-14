package circuit

import (
	"testing"
	"time"
)

var fixedNow = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func makeStore(threshold int) *Store {
	return New(threshold)
}

func TestRecordMiss_BelowThreshold(t *testing.T) {
	s := makeStore(3)
	justOpened := s.RecordMiss("backup", fixedNow)
	if justOpened {
		t.Fatal("circuit should not open on first miss")
	}
	if s.IsOpen("backup") {
		t.Fatal("circuit should still be closed")
	}
}

func TestRecordMiss_OpensAtThreshold(t *testing.T) {
	s := makeStore(2)
	s.RecordMiss("backup", fixedNow)
	justOpened := s.RecordMiss("backup", fixedNow)
	if !justOpened {
		t.Fatal("expected justOpened=true on threshold hit")
	}
	if !s.IsOpen("backup") {
		t.Fatal("circuit should be open")
	}
}

func TestRecordMiss_AlreadyOpen_NoDoubleOpen(t *testing.T) {
	s := makeStore(2)
	s.RecordMiss("backup", fixedNow)
	s.RecordMiss("backup", fixedNow) // opens
	justOpened := s.RecordMiss("backup", fixedNow)
	if justOpened {
		t.Fatal("justOpened should be false when circuit was already open")
	}
}

func TestReset_ClosesCircuit(t *testing.T) {
	s := makeStore(1)
	s.RecordMiss("db-check", fixedNow)
	if !s.IsOpen("db-check") {
		t.Fatal("circuit should be open before reset")
	}
	if err := s.Reset("db-check"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.IsOpen("db-check") {
		t.Fatal("circuit should be closed after reset")
	}
	st, _ := s.Get("db-check")
	if st.ConsecutiveMisses != 0 {
		t.Fatalf("expected ConsecutiveMisses=0, got %d", st.ConsecutiveMisses)
	}
}

func TestReset_UnknownJob(t *testing.T) {
	s := makeStore(3)
	if err := s.Reset("ghost"); err == nil {
		t.Fatal("expected error for unknown job")
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := makeStore(3)
	_, ok := s.Get("missing")
	if ok {
		t.Fatal("expected ok=false for unknown job")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := makeStore(5)
	s.RecordMiss("job-a", fixedNow)
	s.RecordMiss("job-b", fixedNow)
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestDefaultThreshold(t *testing.T) {
	s := New(0) // should default to 3
	for i := 0; i < 2; i++ {
		s.RecordMiss("x", fixedNow)
	}
	if s.IsOpen("x") {
		t.Fatal("circuit should not open before default threshold of 3")
	}
	s.RecordMiss("x", fixedNow)
	if !s.IsOpen("x") {
		t.Fatal("circuit should open at default threshold of 3")
	}
}
