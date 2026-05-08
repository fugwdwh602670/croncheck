package fingerprint

import (
	"testing"
	"time"
)

var epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func TestCompute_Deterministic(t *testing.T) {
	a := Compute("backup", "missed")
	b := Compute("backup", "missed")
	if a != b {
		t.Fatalf("expected same fingerprint, got %s vs %s", a, b)
	}
}

func TestCompute_DifferentInputs(t *testing.T) {
	a := Compute("backup", "missed")
	b := Compute("backup", "failed")
	if a == b {
		t.Fatal("expected different fingerprints for different reasons")
	}
}

func TestIsDuplicate_FirstCall_ReturnsFalse(t *testing.T) {
	s := New()
	if s.IsDuplicate("job1", "missed", epoch) {
		t.Fatal("first call should not be a duplicate")
	}
}

func TestIsDuplicate_SecondCall_ReturnsTrue(t *testing.T) {
	s := New()
	s.IsDuplicate("job1", "missed", epoch)
	if !s.IsDuplicate("job1", "missed", epoch.Add(time.Minute)) {
		t.Fatal("second call should be a duplicate")
	}
}

func TestIsDuplicate_UpdatesCount(t *testing.T) {
	s := New()
	s.IsDuplicate("job1", "missed", epoch)
	s.IsDuplicate("job1", "missed", epoch.Add(time.Minute))

	fp := Compute("job1", "missed")
	e, ok := s.Get(fp)
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Count != 2 {
		t.Fatalf("expected count 2, got %d", e.Count)
	}
}

func TestIsDuplicate_IndependentJobs(t *testing.T) {
	s := New()
	s.IsDuplicate("job1", "missed", epoch)
	if s.IsDuplicate("job2", "missed", epoch) {
		t.Fatal("different jobs should not share fingerprints")
	}
}

func TestGet_UnknownFingerprint(t *testing.T) {
	s := New()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected false for unknown fingerprint")
	}
}

func TestRemove_AllowsReentry(t *testing.T) {
	s := New()
	s.IsDuplicate("job1", "missed", epoch)
	fp := Compute("job1", "missed")
	s.Remove(fp)
	if s.IsDuplicate("job1", "missed", epoch.Add(time.Hour)) {
		t.Fatal("after Remove, next call should not be a duplicate")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	s.IsDuplicate("job1", "missed", epoch)
	s.IsDuplicate("job2", "failed", epoch)
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}
