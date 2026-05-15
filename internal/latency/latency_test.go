package latency

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	current := t
	return func() time.Time { return current }
}

func makeStore(t time.Time) (*Store, *time.Time) {
	current := t
	s := New(func() time.Time { return current })
	return s, &current
}

func TestRecord_FirstHeartbeat_NoSample(t *testing.T) {
	base := time.Now()
	s, _ := makeStore(base)
	if err := s.Record("backup"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, err := s.Get("backup")
	if err != nil {
		t.Fatalf("expected entry after first record: %v", err)
	}
	if e.SampleCount != 0 {
		t.Errorf("expected 0 samples after first heartbeat, got %d", e.SampleCount)
	}
}

func TestRecord_SecondHeartbeat_ProducesSample(t *testing.T) {
	base := time.Now()
	current := base
	s := New(func() time.Time { return current })

	_ = s.Record("backup")
	current = base.Add(5 * time.Minute)
	_ = s.Record("backup")

	e, err := s.Get("backup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.SampleCount != 1 {
		t.Errorf("expected 1 sample, got %d", e.SampleCount)
	}
	if e.Max != 5*time.Minute {
		t.Errorf("expected max 5m, got %v", e.Max)
	}
	if e.Min != 5*time.Minute {
		t.Errorf("expected min 5m, got %v", e.Min)
	}
}

func TestRecord_TracksMinMax(t *testing.T) {
	base := time.Now()
	current := base
	s := New(func() time.Time { return current })

	_ = s.Record("job")
	current = base.Add(3 * time.Minute)
	_ = s.Record("job")
	current = current.Add(7 * time.Minute)
	_ = s.Record("job")

	e, _ := s.Get("job")
	if e.Min != 3*time.Minute {
		t.Errorf("expected min 3m, got %v", e.Min)
	}
	if e.Max != 7*time.Minute {
		t.Errorf("expected max 7m, got %v", e.Max)
	}
	if e.SampleCount != 2 {
		t.Errorf("expected 2 samples, got %d", e.SampleCount)
	}
}

func TestRecord_EmptyJobReturnsError(t *testing.T) {
	s := New(nil)
	if err := s.Record(""); err == nil {
		t.Error("expected error for empty job name")
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New(nil)
	_, err := s.Get("ghost")
	if err == nil {
		t.Error("expected error for unknown job")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	base := time.Now()
	current := base
	s := New(func() time.Time { return current })

	_ = s.Record("job-a")
	_ = s.Record("job-b")
	current = base.Add(1 * time.Minute)
	_ = s.Record("job-a")

	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestAll_EmptyStore(t *testing.T) {
	s := New(nil)
	if got := s.All(); len(got) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(got))
	}
}
