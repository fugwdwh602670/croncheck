package sla

import (
	"testing"
)

func TestRecordCheck_FirstCheck_Hit(t *testing.T) {
	s := New()
	s.RecordCheck("backup", false)

	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Total != 1 {
		t.Errorf("expected Total=1, got %d", e.Total)
	}
	if e.Missed != 0 {
		t.Errorf("expected Missed=0, got %d", e.Missed)
	}
	if e.Compliance != 100.0 {
		t.Errorf("expected Compliance=100, got %.2f", e.Compliance)
	}
}

func TestRecordCheck_Missed(t *testing.T) {
	s := New()
	s.RecordCheck("backup", false)
	s.RecordCheck("backup", false)
	s.RecordCheck("backup", true)

	e, _ := s.Get("backup")
	if e.Total != 3 {
		t.Errorf("expected Total=3, got %d", e.Total)
	}
	if e.Missed != 1 {
		t.Errorf("expected Missed=1, got %d", e.Missed)
	}
	want := 66.66666666666667
	if e.Compliance < 66.0 || e.Compliance > 67.0 {
		t.Errorf("expected Compliance~=%.2f, got %.2f", want, e.Compliance)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("unknown")
	if ok {
		t.Error("expected ok=false for unknown job")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	s.RecordCheck("job-a", false)
	s.RecordCheck("job-b", true)

	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestReset_ClearsEntry(t *testing.T) {
	s := New()
	s.RecordCheck("nightly", false)
	s.Reset("nightly")

	_, ok := s.Get("nightly")
	if ok {
		t.Error("expected entry to be cleared after Reset")
	}
}

func TestCompliance_ZeroTotal(t *testing.T) {
	got := compliance(0, 0)
	if got != 100.0 {
		t.Errorf("expected 100.0 for zero total, got %.2f", got)
	}
}
