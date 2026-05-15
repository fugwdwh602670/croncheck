package drift

import (
	"testing"
	"time"
)

var epoch = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func makeStore() *Store {
	s := New()
	s.now = func() time.Time { return epoch }
	return s
}

func TestRecord_StoresDrift(t *testing.T) {
	s := makeStore()
	expected := epoch
	actual := epoch.Add(5 * time.Second)
	if err := s.Record("backup", expected, actual); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry, got none")
	}
	if e.LastDrift != 5*time.Second {
		t.Errorf("LastDrift = %v, want 5s", e.LastDrift)
	}
	if e.SampleCount != 1 {
		t.Errorf("SampleCount = %d, want 1", e.SampleCount)
	}
}

func TestRecord_TracksMaxDrift(t *testing.T) {
	s := makeStore()
	s.Record("job", epoch, epoch.Add(3*time.Second))
	s.Record("job", epoch, epoch.Add(10*time.Second))
	s.Record("job", epoch, epoch.Add(2*time.Second))
	e, _ := s.Get("job")
	if e.MaxDrift != 10*time.Second {
		t.Errorf("MaxDrift = %v, want 10s", e.MaxDrift)
	}
}

func TestRecord_ComputesAvg(t *testing.T) {
	s := makeStore()
	s.Record("job", epoch, epoch.Add(4*time.Second))
	s.Record("job", epoch, epoch.Add(8*time.Second))
	s.Record("job", epoch, epoch.Add(6*time.Second))
	e, _ := s.Get("job")
	if e.AvgDrift != 6*time.Second {
		t.Errorf("AvgDrift = %v, want 6s", e.AvgDrift)
	}
}

func TestRecord_NegativeDeltaNormalised(t *testing.T) {
	s := makeStore()
	// actual before expected — still treated as positive drift
	s.Record("early", epoch, epoch.Add(-3*time.Second))
	e, _ := s.Get("early")
	if e.LastDrift != 3*time.Second {
		t.Errorf("LastDrift = %v, want 3s", e.LastDrift)
	}
}

func TestRecord_EmptyJobReturnsError(t *testing.T) {
	s := makeStore()
	if err := s.Record("", epoch, epoch.Add(time.Second)); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := makeStore()
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected false for unknown job")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := makeStore()
	s.Record("a", epoch, epoch.Add(time.Second))
	s.Record("b", epoch, epoch.Add(2*time.Second))
	all := s.All()
	if len(all) != 2 {
		t.Errorf("All() len = %d, want 2", len(all))
	}
}

func TestReset_ClearsEntry(t *testing.T) {
	s := makeStore()
	s.Record("job", epoch, epoch.Add(time.Second))
	s.Reset("job")
	_, ok := s.Get("job")
	if ok {
		t.Fatal("expected entry to be cleared after Reset")
	}
}
