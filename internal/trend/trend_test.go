package trend

import (
	"testing"
	"time"
)

var fixedNow = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func makeStore() *Store {
	s := New(10 * time.Minute)
	s.now = func() time.Time { return fixedNow }
	return s
}

func TestRecord_StoresHit(t *testing.T) {
	s := makeStore()
	if err := s.Record("backup", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry")
	}
	if e.Total != 1 || e.Misses != 0 {
		t.Errorf("got total=%d misses=%d", e.Total, e.Misses)
	}
}

func TestRecord_StoresMiss(t *testing.T) {
	s := makeStore()
	_ = s.Record("backup", true)
	e, _ := s.Get("backup")
	if e.Misses != 1 || e.MissRate != 1.0 {
		t.Errorf("got misses=%d rate=%f", e.Misses, e.MissRate)
	}
}

func TestRecord_EmptyJobReturnsError(t *testing.T) {
	s := makeStore()
	if err := s.Record("", false); err == nil {
		t.Fatal("expected error for empty job")
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := makeStore()
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestMissRate_MixedEvents(t *testing.T) {
	s := makeStore()
	_ = s.Record("job", false)
	_ = s.Record("job", true)
	_ = s.Record("job", false)
	_ = s.Record("job", true)
	e, _ := s.Get("job")
	if e.Total != 4 || e.Misses != 2 {
		t.Errorf("got total=%d misses=%d", e.Total, e.Misses)
	}
	if e.MissRate != 0.5 {
		t.Errorf("expected rate 0.5, got %f", e.MissRate)
	}
}

func TestPrune_RemovesOldEvents(t *testing.T) {
	s := New(5 * time.Minute)
	old := fixedNow.Add(-10 * time.Minute)
	s.now = func() time.Time { return old }
	_ = s.Record("job", true)
	s.now = func() time.Time { return fixedNow }
	_ = s.Record("job", false)
	e, _ := s.Get("job")
	if e.Total != 1 {
		t.Errorf("expected 1 event after pruning, got %d", e.Total)
	}
	if e.Misses != 0 {
		t.Errorf("expected 0 misses after pruning, got %d", e.Misses)
	}
}

func TestAll_ReturnsAllJobs(t *testing.T) {
	s := makeStore()
	_ = s.Record("a", false)
	_ = s.Record("b", true)
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}
