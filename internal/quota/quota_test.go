package quota

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time { return func() time.Time { return t } }

func makeStore(t *testing.T, max int, window time.Duration) *Store {
	t.Helper()
	s, err := New(max, window)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

func TestAllow_WithinQuota(t *testing.T) {
	s := makeStore(t, 3, time.Minute)
	for i := 0; i < 3; i++ {
		if !s.Allow("job1") {
			t.Fatalf("expected Allow=true on call %d", i+1)
		}
	}
}

func TestAllow_ExceedsQuota(t *testing.T) {
	s := makeStore(t, 2, time.Minute)
	s.Allow("job1")
	s.Allow("job1")
	if s.Allow("job1") {
		t.Fatal("expected Allow=false after quota exhausted")
	}
}

func TestAllow_ResetsAfterWindow(t *testing.T) {
	base := time.Now()
	s := makeStore(t, 1, time.Minute)
	s.now = fixedNow(base)
	s.Allow("job1")
	if s.Allow("job1") {
		t.Fatal("expected Allow=false within window")
	}
	s.now = fixedNow(base.Add(2 * time.Minute))
	if !s.Allow("job1") {
		t.Fatal("expected Allow=true after window expired")
	}
}

func TestAllow_IndependentJobs(t *testing.T) {
	s := makeStore(t, 1, time.Minute)
	s.Allow("job1")
	if !s.Allow("job2") {
		t.Fatal("expected job2 to be unaffected by job1 quota")
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := makeStore(t, 5, time.Minute)
	if s.Get("missing") != nil {
		t.Fatal("expected nil for unknown job")
	}
}

func TestGet_ReturnsSnapshot(t *testing.T) {
	s := makeStore(t, 5, time.Minute)
	s.Allow("job1")
	e := s.Get("job1")
	if e == nil || e.Count != 1 {
		t.Fatalf("expected count=1, got %v", e)
	}
}

func TestReset_AllowsImmediately(t *testing.T) {
	s := makeStore(t, 1, time.Minute)
	s.Allow("job1")
	s.Reset("job1")
	if !s.Allow("job1") {
		t.Fatal("expected Allow=true after reset")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := makeStore(t, 5, time.Minute)
	s.Allow("a")
	s.Allow("b")
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestNew_InvalidMax(t *testing.T) {
	_, err := New(0, time.Minute)
	if err == nil {
		t.Fatal("expected error for max=0")
	}
}

func TestNew_InvalidWindow(t *testing.T) {
	_, err := New(1, 0)
	if err == nil {
		t.Fatal("expected error for window=0")
	}
}
