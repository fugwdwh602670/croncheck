package ack

import (
	"testing"
	"time"
)

func TestIsAcked_ActiveAck(t *testing.T) {
	s := New()
	s.Add("backup", "investigating", time.Hour)
	if !s.IsAcked("backup") {
		t.Fatal("expected job to be acked")
	}
}

func TestIsAcked_ExpiredAck(t *testing.T) {
	s := New()
	fixed := time.Now()
	s.now = func() time.Time { return fixed }
	s.Add("backup", "", time.Millisecond)
	s.now = func() time.Time { return fixed.Add(time.Second) }
	if s.IsAcked("backup") {
		t.Fatal("expected ack to be expired")
	}
}

func TestIsAcked_UnknownJob(t *testing.T) {
	s := New()
	if s.IsAcked("unknown") {
		t.Fatal("expected unknown job to not be acked")
	}
}

func TestRemove_ClearsAck(t *testing.T) {
	s := New()
	s.Add("backup", "", time.Hour)
	s.Remove("backup")
	if s.IsAcked("backup") {
		t.Fatal("expected ack to be removed")
	}
}

func TestAll_ReturnsOnlyActive(t *testing.T) {
	s := New()
	fixed := time.Now()
	s.now = func() time.Time { return fixed }
	s.Add("job-a", "reason", time.Hour)
	s.Add("job-b", "", time.Millisecond)
	s.now = func() time.Time { return fixed.Add(time.Second) }
	all := s.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 active ack, got %d", len(all))
	}
	if all[0].JobName != "job-a" {
		t.Errorf("unexpected job name %q", all[0].JobName)
	}
}

func TestAll_ReasonPreserved(t *testing.T) {
	s := New()
	s.Add("deploy", "planned maintenance", time.Hour)
	all := s.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 ack, got %d", len(all))
	}
	if all[0].Reason != "planned maintenance" {
		t.Errorf("unexpected reason %q", all[0].Reason)
	}
}
