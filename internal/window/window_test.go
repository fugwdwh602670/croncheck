package window

import (
	"testing"
	"time"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	if err := s.Set("backup", 0, 5*time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Job != "backup" || e.Start != 0 || e.End != 5*time.Minute {
		t.Errorf("unexpected entry: %+v", e)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestSet_EmptyJobReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("", 0, time.Minute); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestSet_EndNotGreaterThanStartReturnsError(t *testing.T) {
	s := New()
	if err := s.Set("job", 5*time.Minute, 5*time.Minute); err == nil {
		t.Fatal("expected error when end == start")
	}
	if err := s.Set("job", 10*time.Minute, 5*time.Minute); err == nil {
		t.Fatal("expected error when end < start")
	}
}

func TestRemove_ClearsEntry(t *testing.T) {
	s := New()
	_ = s.Set("cleanup", 0, time.Minute)
	s.Remove("cleanup")
	_, ok := s.Get("cleanup")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Set("a", 0, time.Minute)
	_ = s.Set("b", 0, 2*time.Minute)
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestInWindow_WithinBounds(t *testing.T) {
	s := New()
	_ = s.Set("sync", 0, 10*time.Minute)
	sched := time.Now().Truncate(time.Hour)
	if !s.InWindow("sync", sched, sched.Add(5*time.Minute)) {
		t.Error("expected in-window result")
	}
}

func TestInWindow_OutsideBounds(t *testing.T) {
	s := New()
	_ = s.Set("sync", 0, 10*time.Minute)
	sched := time.Now().Truncate(time.Hour)
	if s.InWindow("sync", sched, sched.Add(15*time.Minute)) {
		t.Error("expected out-of-window result")
	}
}

func TestInWindow_NoConfigPermissive(t *testing.T) {
	s := New()
	sched := time.Now()
	if !s.InWindow("unconfigured", sched, sched.Add(time.Hour)) {
		t.Error("expected permissive result when no window configured")
	}
}
