package forecast

import (
	"testing"
	"time"
)

func makeStore() *Store { return New(5) }

func TestRecord_FirstHeartbeat_NoForecast(t *testing.T) {
	s := makeStore()
	now := time.Now()
	if err := s.Record("backup", now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := s.Get("backup")
	if ok {
		t.Fatal("expected no forecast after first heartbeat")
	}
}

func TestRecord_SecondHeartbeat_ProducesForecast(t *testing.T) {
	s := makeStore()
	base := time.Now()
	_ = s.Record("backup", base)
	_ = s.Record("backup", base.Add(10*time.Minute))

	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected forecast after second heartbeat")
	}
	if e.AvgInterval != 10*time.Minute {
		t.Errorf("avg interval = %v, want 10m", e.AvgInterval)
	}
	if e.SampleCount != 1 {
		t.Errorf("sample count = %d, want 1", e.SampleCount)
	}
	want := base.Add(10 * time.Minute).Add(10 * time.Minute)
	if !e.NextExpected.Equal(want) {
		t.Errorf("next expected = %v, want %v", e.NextExpected, want)
	}
}

func TestRecord_AveragesMultipleSamples(t *testing.T) {
	s := makeStore()
	base := time.Now()
	_ = s.Record("job", base)
	_ = s.Record("job", base.Add(10*time.Minute))
	_ = s.Record("job", base.Add(10*time.Minute+12*time.Minute))

	e, _ := s.Get("job")
	// intervals: 10m, 12m → avg 11m
	if e.AvgInterval != 11*time.Minute {
		t.Errorf("avg = %v, want 11m", e.AvgInterval)
	}
	if e.SampleCount != 2 {
		t.Errorf("sample count = %d, want 2", e.SampleCount)
	}
}

func TestRecord_EmptyJobReturnsError(t *testing.T) {
	s := makeStore()
	if err := s.Record("", time.Now()); err == nil {
		t.Fatal("expected error for empty job")
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := makeStore()
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected false for unknown job")
	}
}

func TestAll_ReturnsOnlyForecasted(t *testing.T) {
	s := makeStore()
	base := time.Now()
	_ = s.Record("a", base)
	_ = s.Record("b", base)
	_ = s.Record("a", base.Add(5*time.Minute))

	all := s.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 forecast, got %d", len(all))
	}
	if all[0].Job != "a" {
		t.Errorf("expected job a, got %s", all[0].Job)
	}
}

func TestRecord_RespectsMaxSamples(t *testing.T) {
	s := New(3)
	base := time.Now()
	_ = s.Record("job", base)
	for i := 1; i <= 6; i++ {
		_ = s.Record("job", base.Add(time.Duration(i)*time.Minute))
	}
	e, _ := s.Get("job")
	if e.SampleCount != 3 {
		t.Errorf("sample count = %d, want 3 (capped)", e.SampleCount)
	}
}
