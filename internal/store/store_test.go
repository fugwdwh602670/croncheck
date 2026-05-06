package store

import (
	"testing"
	"time"
)

func TestRecordHeartbeat_NewJob(t *testing.T) {
	s := New()
	before := time.Now()
	s.RecordHeartbeat("backup")
	after := time.Now()

	js, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected job to exist after heartbeat")
	}
	if js.LastSeen.Before(before) || js.LastSeen.After(after) {
		t.Errorf("LastSeen %v not in expected range [%v, %v]", js.LastSeen, before, after)
	}
	if js.MissedCount != 0 {
		t.Errorf("expected MissedCount 0, got %d", js.MissedCount)
	}
}

func TestRecordHeartbeat_ResetsState(t *testing.T) {
	s := New()
	s.RecordHeartbeat("backup")
	s.IncrementMissed("backup")

	js, _ := s.Get("backup")
	if js.MissedCount != 1 || !js.Alerted {
		t.Fatal("pre-condition: job should be missed and alerted")
	}

	s.RecordHeartbeat("backup")
	js, _ = s.Get("backup")
	if js.MissedCount != 0 {
		t.Errorf("expected MissedCount reset to 0, got %d", js.MissedCount)
	}
	if js.Alerted {
		t.Error("expected Alerted to be reset to false")
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("expected false for unknown job")
	}
}

func TestIncrementMissed(t *testing.T) {
	s := New()
	s.RecordHeartbeat("sync")
	s.IncrementMissed("sync")
	s.IncrementMissed("sync")

	js, ok := s.Get("sync")
	if !ok {
		t.Fatal("expected job to exist")
	}
	if js.MissedCount != 2 {
		t.Errorf("expected MissedCount 2, got %d", js.MissedCount)
	}
	if !js.Alerted {
		t.Error("expected Alerted to be true")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	s.RecordHeartbeat("job-a")
	s.RecordHeartbeat("job-b")

	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(all))
	}
}

func TestAll_EmptyStore(t *testing.T) {
	s := New()
	all := s.All()
	if len(all) != 0 {
		t.Errorf("expected 0 jobs for empty store, got %d", len(all))
	}
}

func TestIncrementMissed_UnknownJob(t *testing.T) {
	s := New()
	// Should not panic for unknown job
	s.IncrementMissed("ghost")
}
