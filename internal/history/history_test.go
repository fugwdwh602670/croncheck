package history

import (
	"testing"
	"time"
)

func makeEvent(job, kind string) Event {
	return Event{JobName: job, Kind: kind, OccuredAt: time.Now()}
}

func TestRecord_SingleEvent(t *testing.T) {
	r := New(5)
	r.Record(makeEvent("backup", "missed"))

	events := r.Get("backup")
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Kind != "missed" {
		t.Errorf("expected kind 'missed', got %q", events[0].Kind)
	}
}

func TestRecord_RespectsLimit(t *testing.T) {
	const limit = 3
	r := New(limit)

	for i := 0; i < 5; i++ {
		r.Record(makeEvent("sync", "missed"))
	}

	events := r.Get("sync")
	if len(events) != limit {
		t.Fatalf("expected %d events after overflow, got %d", limit, len(events))
	}
}

func TestRecord_DefaultLimit(t *testing.T) {
	r := New(0) // should default to 10
	for i := 0; i < 15; i++ {
		r.Record(makeEvent("job", "failed"))
	}
	if len(r.Get("job")) != 10 {
		t.Errorf("expected default limit of 10")
	}
}

func TestGet_UnknownJob(t *testing.T) {
	r := New(5)
	if r.Get("nonexistent") != nil {
		t.Error("expected nil for unknown job")
	}
}

func TestAll_ReturnsAllJobs(t *testing.T) {
	r := New(5)
	r.Record(makeEvent("alpha", "missed"))
	r.Record(makeEvent("beta", "failed"))
	r.Record(makeEvent("alpha", "missed"))

	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(all))
	}
	if len(all["alpha"]) != 2 {
		t.Errorf("expected 2 events for alpha, got %d", len(all["alpha"]))
	}
	if len(all["beta"]) != 1 {
		t.Errorf("expected 1 event for beta, got %d", len(all["beta"]))
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	r := New(5)
	r.Record(makeEvent("job", "missed"))

	all := r.All()
	all["job"][0].Kind = "mutated"

	fresh := r.Get("job")
	if fresh[0].Kind == "mutated" {
		t.Error("All() returned a reference to internal slice")
	}
}
