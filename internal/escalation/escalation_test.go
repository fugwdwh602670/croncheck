package escalation

import (
	"testing"
	"time"
)

var fixedNow = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func TestRecord_NormalLevel(t *testing.T) {
	s := New(3, 5)
	e := s.Record("job-a", fixedNow)
	if e.FireCount != 1 {
		t.Fatalf("expected FireCount=1, got %d", e.FireCount)
	}
	if e.Level != LevelNormal {
		t.Fatalf("expected LevelNormal, got %d", e.Level)
	}
}

func TestRecord_WarningLevel(t *testing.T) {
	s := New(3, 5)
	var e Entry
	for i := 0; i < 3; i++ {
		e = s.Record("job-a", fixedNow)
	}
	if e.Level != LevelWarning {
		t.Fatalf("expected LevelWarning, got %d", e.Level)
	}
	if e.EscalatedAt.IsZero() {
		t.Fatal("expected EscalatedAt to be set on warning transition")
	}
}

func TestRecord_CriticalLevel(t *testing.T) {
	s := New(3, 5)
	var e Entry
	for i := 0; i < 5; i++ {
		e = s.Record("job-a", fixedNow)
	}
	if e.Level != LevelCritical {
		t.Fatalf("expected LevelCritical, got %d", e.Level)
	}
}

func TestRecord_IndependentJobs(t *testing.T) {
	s := New(3, 5)
	for i := 0; i < 4; i++ {
		s.Record("job-a", fixedNow)
	}
	e := s.Record("job-b", fixedNow)
	if e.FireCount != 1 {
		t.Fatalf("job-b should have FireCount=1, got %d", e.FireCount)
	}
	if e.Level != LevelNormal {
		t.Fatalf("job-b should be LevelNormal, got %d", e.Level)
	}
}

func TestReset_ClearsEntry(t *testing.T) {
	s := New(3, 5)
	for i := 0; i < 4; i++ {
		s.Record("job-a", fixedNow)
	}
	s.Reset("job-a")
	_, ok := s.Get("job-a")
	if ok {
		t.Fatal("expected entry to be cleared after Reset")
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New(3, 5)
	_, ok := s.Get("unknown")
	if ok {
		t.Fatal("expected ok=false for unknown job")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New(3, 5)
	s.Record("job-a", fixedNow)
	s.Record("job-b", fixedNow)
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// Mutating the snapshot must not affect the store.
	all["job-a"] = Entry{FireCount: 999}
	e, _ := s.Get("job-a")
	if e.FireCount == 999 {
		t.Fatal("snapshot mutation affected internal store")
	}
}
