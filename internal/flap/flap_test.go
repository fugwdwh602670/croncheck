package flap

import (
	"testing"
	"time"
)

var epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func TestRecord_NotFlappingBelowThreshold(t *testing.T) {
	s := New(10*time.Minute, 4)
	for i := 0; i < 3; i++ {
		flapping := s.Record("job1", epoch.Add(time.Duration(i)*time.Minute))
		if flapping {
			t.Fatalf("expected not flapping after %d changes", i+1)
		}
	}
}

func TestRecord_FlappingAtThreshold(t *testing.T) {
	s := New(10*time.Minute, 4)
	var result bool
	for i := 0; i < 4; i++ {
		result = s.Record("job1", epoch.Add(time.Duration(i)*time.Minute))
	}
	if !result {
		t.Fatal("expected flapping after reaching threshold")
	}
}

func TestRecord_ResetsAfterWindow(t *testing.T) {
	s := New(5*time.Minute, 3)
	// Record changes within window
	s.Record("job1", epoch)
	s.Record("job1", epoch.Add(1*time.Minute))
	// Next change is outside the window — counter should reset
	flapping := s.Record("job1", epoch.Add(10*time.Minute))
	if flapping {
		t.Fatal("expected counter reset after window elapsed")
	}
	e := s.entries["job1"]
	if e.Changes != 1 {
		t.Fatalf("expected Changes=1 after reset, got %d", e.Changes)
	}
}

func TestIsFlapping_UnknownJob(t *testing.T) {
	s := New(10*time.Minute, 4)
	if s.IsFlapping("ghost") {
		t.Fatal("expected false for unknown job")
	}
}

func TestIsFlapping_AfterRecord(t *testing.T) {
	s := New(10*time.Minute, 2)
	s.Record("job1", epoch)
	s.Record("job1", epoch.Add(time.Minute))
	if !s.IsFlapping("job1") {
		t.Fatal("expected IsFlapping to return true")
	}
}

func TestReset_ClearsEntry(t *testing.T) {
	s := New(10*time.Minute, 2)
	s.Record("job1", epoch)
	s.Record("job1", epoch.Add(time.Minute))
	s.Reset("job1")
	if s.IsFlapping("job1") {
		t.Fatal("expected not flapping after reset")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New(10*time.Minute, 4)
	s.Record("a", epoch)
	s.Record("b", epoch)
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if _, ok := all["a"]; !ok {
		t.Error("missing entry for 'a'")
	}
	if _, ok := all["b"]; !ok {
		t.Error("missing entry for 'b'")
	}
}

func TestAll_IsolatesMutation(t *testing.T) {
	s := New(10*time.Minute, 4)
	s.Record("job1", epoch)
	all := s.All()
	all["job1"] = Entry{Changes: 999}
	if s.entries["job1"].Changes == 999 {
		t.Fatal("All snapshot should not share memory with internal state")
	}
}
