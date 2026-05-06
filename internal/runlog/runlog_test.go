package runlog

import (
	"testing"
	"time"
)

func TestRecord_StoresEntry(t *testing.T) {
	s := New()
	before := time.Now().UTC()
	s.Record("backup", StatusSuccess, "exit 0")

	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to be present")
	}
	if e.JobName != "backup" {
		t.Errorf("got JobName %q, want %q", e.JobName, "backup")
	}
	if e.Status != StatusSuccess {
		t.Errorf("got Status %q, want %q", e.Status, StatusSuccess)
	}
	if e.Message != "exit 0" {
		t.Errorf("got Message %q, want %q", e.Message, "exit 0")
	}
	if e.RecordedAt.Before(before) {
		t.Error("RecordedAt should not be before the test start")
	}
}

func TestRecord_OverwritesPreviousEntry(t *testing.T) {
	s := New()
	s.Record("deploy", StatusSuccess, "ok")
	s.Record("deploy", StatusFailure, "non-zero exit")

	e, _ := s.Get("deploy")
	if e.Status != StatusFailure {
		t.Errorf("expected overwritten status %q, got %q", StatusFailure, e.Status)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("ghost")
	if ok {
		t.Error("expected ok=false for unknown job")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	s.Record("job-a", StatusSuccess, "")
	s.Record("job-b", StatusFailure, "timeout")

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestAll_EmptyStore(t *testing.T) {
	s := New()
	if got := s.All(); len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}

func TestClear_RemovesEntry(t *testing.T) {
	s := New()
	s.Record("cleanup", StatusSuccess, "done")
	s.Clear("cleanup")

	_, ok := s.Get("cleanup")
	if ok {
		t.Error("expected entry to be removed after Clear")
	}
}

func TestClear_NoopForUnknownJob(t *testing.T) {
	s := New()
	// Should not panic.
	s.Clear("nonexistent")
}
