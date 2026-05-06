package inhibit

import (
	"testing"
	"time"
)

func makeStore() *Store {
	return New([]Rule{
		{SourceJob: "db-backup", TargetJob: "db-report"},
		{SourceJob: "db-backup", TargetJob: "db-cleanup"},
	})
}

func TestIsInhibited_NoUnhealthySources(t *testing.T) {
	s := makeStore()
	if s.IsInhibited("db-report") {
		t.Error("expected not inhibited when source is healthy")
	}
}

func TestIsInhibited_ActiveInhibition(t *testing.T) {
	s := makeStore()
	s.SetUnhealthy("db-backup")
	if !s.IsInhibited("db-report") {
		t.Error("expected db-report to be inhibited")
	}
	if !s.IsInhibited("db-cleanup") {
		t.Error("expected db-cleanup to be inhibited")
	}
}

func TestIsInhibited_UnrelatedJob(t *testing.T) {
	s := makeStore()
	s.SetUnhealthy("db-backup")
	if s.IsInhibited("unrelated-job") {
		t.Error("unrelated job should not be inhibited")
	}
}

func TestSetHealthy_ClearsInhibition(t *testing.T) {
	s := makeStore()
	s.SetUnhealthy("db-backup")
	s.SetHealthy("db-backup")
	if s.IsInhibited("db-report") {
		t.Error("expected inhibition cleared after SetHealthy")
	}
}

func TestSetUnhealthy_RecordsTime(t *testing.T) {
	s := makeStore()
	before := time.Now()
	s.SetUnhealthy("db-backup")
	after := time.Now()

	sources := s.UnhealthySources()
	t0, ok := sources["db-backup"]
	if !ok {
		t.Fatal("expected db-backup in unhealthy sources")
	}
	if t0.Before(before) || t0.After(after) {
		t.Errorf("unexpected timestamp: %v", t0)
	}
}

func TestSetUnhealthy_DoesNotOverwriteTime(t *testing.T) {
	s := makeStore()
	s.SetUnhealthy("db-backup")
	t1 := s.UnhealthySources()["db-backup"]
	time.Sleep(2 * time.Millisecond)
	s.SetUnhealthy("db-backup") // second call should not overwrite
	t2 := s.UnhealthySources()["db-backup"]
	if !t1.Equal(t2) {
		t.Error("SetUnhealthy should not overwrite existing timestamp")
	}
}

func TestRules_ReturnsCopy(t *testing.T) {
	s := makeStore()
	rules := s.Rules()
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	rules[0].SourceJob = "tampered"
	if s.Rules()[0].SourceJob == "tampered" {
		t.Error("Rules() should return a copy")
	}
}
