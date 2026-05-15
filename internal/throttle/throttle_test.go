package throttle

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time { return func() time.Time { return t } }

func makeStore(now time.Time) *Store {
	s := New()
	s.nowFunc = fixedNow(now)
	return s
}

func TestAllow_NoPolicy_AlwaysTrue(t *testing.T) {
	s := makeStore(time.Now())
	if !s.Allow("job") {
		t.Fatal("expected Allow=true when no policy set")
	}
}

func TestAllow_WithinLimit(t *testing.T) {
	now := time.Now()
	s := makeStore(now)
	_ = s.Set("job", Config{MaxAlerts: 3, Window: time.Minute})
	for i := 0; i < 3; i++ {
		if !s.Allow("job") {
			t.Fatalf("expected Allow=true on call %d", i+1)
		}
	}
}

func TestAllow_ExceedsLimit(t *testing.T) {
	now := time.Now()
	s := makeStore(now)
	_ = s.Set("job", Config{MaxAlerts: 2, Window: time.Minute})
	s.Allow("job")
	s.Allow("job")
	if s.Allow("job") {
		t.Fatal("expected Allow=false after limit exceeded")
	}
}

func TestAllow_ResetsAfterWindow(t *testing.T) {
	now := time.Now()
	s := makeStore(now)
	_ = s.Set("job", Config{MaxAlerts: 1, Window: time.Second})
	s.Allow("job")
	if s.Allow("job") {
		t.Fatal("expected Allow=false within window")
	}
	s.nowFunc = fixedNow(now.Add(2 * time.Second))
	if !s.Allow("job") {
		t.Fatal("expected Allow=true after window reset")
	}
}

func TestAllow_IndependentJobs(t *testing.T) {
	s := makeStore(time.Now())
	_ = s.Set("a", Config{MaxAlerts: 1, Window: time.Minute})
	_ = s.Set("b", Config{MaxAlerts: 1, Window: time.Minute})
	s.Allow("a")
	if !s.Allow("b") {
		t.Fatal("expected Allow=true for independent job b")
	}
}

func TestSet_InvalidArgs(t *testing.T) {
	s := New()
	if err := s.Set("", Config{MaxAlerts: 1, Window: time.Second}); err == nil {
		t.Error("expected error for empty job")
	}
	if err := s.Set("job", Config{MaxAlerts: 0, Window: time.Second}); err == nil {
		t.Error("expected error for zero MaxAlerts")
	}
	if err := s.Set("job", Config{MaxAlerts: 1, Window: 0}); err == nil {
		t.Error("expected error for zero Window")
	}
}

func TestRemove_ClearsPolicy(t *testing.T) {
	s := makeStore(time.Now())
	_ = s.Set("job", Config{MaxAlerts: 1, Window: time.Minute})
	s.Allow("job")
	s.Remove("job")
	if !s.Allow("job") {
		t.Fatal("expected Allow=true after Remove")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := New()
	_ = s.Set("a", Config{MaxAlerts: 5, Window: time.Minute})
	_ = s.Set("b", Config{MaxAlerts: 2, Window: time.Hour})
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}
