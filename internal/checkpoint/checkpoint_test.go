package checkpoint

import (
	"testing"
	"time"
)

var fixedNow = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func makeStore() *Store {
	s := New()
	s.now = func() time.Time { return fixedNow }
	return s
}

func TestRecord_StoresEntry(t *testing.T) {
	s := makeStore()
	s.Record("backup")

	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.JobName != "backup" {
		t.Errorf("job name: got %q, want %q", e.JobName, "backup")
	}
	if !e.LastOK.Equal(fixedNow) {
		t.Errorf("LastOK: got %v, want %v", e.LastOK, fixedNow)
	}
}

func TestRecord_OverwritesPreviousEntry(t *testing.T) {
	s := makeStore()
	s.Record("backup")

	later := fixedNow.Add(5 * time.Minute)
	s.now = func() time.Time { return later }
	s.Record("backup")

	e, _ := s.Get("backup")
	if !e.LastOK.Equal(later) {
		t.Errorf("expected LastOK to be updated to %v, got %v", later, e.LastOK)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := makeStore()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for unknown job")
	}
}

func TestRemove_ClearsEntry(t *testing.T) {
	s := makeStore()
	s.Record("cleanup")
	s.Remove("cleanup")

	_, ok := s.Get("cleanup")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := makeStore()
	s.Record("jobA")
	s.Record("jobB")

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestAll_EmptyStore(t *testing.T) {
	s := makeStore()
	if got := s.All(); len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}
