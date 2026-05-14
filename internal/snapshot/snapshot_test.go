package snapshot

import (
	"testing"
	"time"
)

type mockSource struct {
	states []JobState
}

func (m *mockSource) AllStates() []JobState {
	out := make([]JobState, len(m.states))
	copy(out, m.states)
	return out
}

func makeStore(limit int) *Store {
	s := New(limit)
	s.now = func() time.Time { return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) }
	return s
}

func TestCapture_StoresSnapshot(t *testing.T) {
	s := makeStore(5)
	src := &mockSource{states: []JobState{{Job: "backup", Healthy: true, MissedCount: 0}}}
	snap := s.Capture("snap-1", src)
	if snap.ID != "snap-1" {
		t.Fatalf("expected ID snap-1, got %s", snap.ID)
	}
	if len(snap.Jobs) != 1 || snap.Jobs[0].Job != "backup" {
		t.Fatalf("unexpected jobs: %+v", snap.Jobs)
	}
}

func TestGet_KnownSnapshot(t *testing.T) {
	s := makeStore(5)
	src := &mockSource{states: []JobState{{Job: "etl", Healthy: false, MissedCount: 2}}}
	s.Capture("snap-a", src)
	snap, ok := s.Get("snap-a")
	if !ok {
		t.Fatal("expected snapshot to be found")
	}
	if snap.Jobs[0].MissedCount != 2 {
		t.Fatalf("expected missed_count 2, got %d", snap.Jobs[0].MissedCount)
	}
}

func TestGet_UnknownSnapshot(t *testing.T) {
	s := makeStore(5)
	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestCapture_RespectsLimit(t *testing.T) {
	s := makeStore(3)
	src := &mockSource{}
	for i := 0; i < 5; i++ {
		s.Capture("id", src)
	}
	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 snapshots, got %d", len(all))
	}
}

func TestAll_NewestFirst(t *testing.T) {
	s := New(5)
	times := []time.Time{
		time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	ids := []string{"first", "second", "third"}
	src := &mockSource{}
	for i, id := range ids {
		t := times[i]
		s.now = func() time.Time { return t }
		s.Capture(id, src)
	}
	all := s.All()
	if all[0].ID != "third" {
		t.Fatalf("expected newest first, got %s", all[0].ID)
	}
}

func TestCapture_DefaultLimit(t *testing.T) {
	s := New(0)
	if s.limit != 10 {
		t.Fatalf("expected default limit 10, got %d", s.limit)
	}
}
