package silence_test

import (
	"testing"
	"time"

	"github.com/example/croncheck/internal/silence"
)

func TestIsSilenced_ActiveSilence(t *testing.T) {
	r := silence.New()
	now := time.Now()
	r.Add("backup", "maintenance", now.Add(time.Hour))

	if !r.IsSilenced("backup", now) {
		t.Fatal("expected job to be silenced")
	}
}

func TestIsSilenced_ExpiredSilence(t *testing.T) {
	r := silence.New()
	now := time.Now()
	r.Add("backup", "old maintenance", now.Add(-time.Minute))

	if r.IsSilenced("backup", now) {
		t.Fatal("expected expired silence to be inactive")
	}
}

func TestIsSilenced_UnknownJob(t *testing.T) {
	r := silence.New()
	if r.IsSilenced("nonexistent", time.Now()) {
		t.Fatal("expected unknown job to not be silenced")
	}
}

func TestRemove_ClearsSilence(t *testing.T) {
	r := silence.New()
	now := time.Now()
	r.Add("deploy", "deploy window", now.Add(time.Hour))
	r.Remove("deploy")

	if r.IsSilenced("deploy", now) {
		t.Fatal("expected silence to be removed")
	}
}

func TestAll_ReturnsOnlyActive(t *testing.T) {
	r := silence.New()
	now := time.Now()
	r.Add("job-a", "active", now.Add(time.Hour))
	r.Add("job-b", "expired", now.Add(-time.Minute))

	all := r.All(now)
	if len(all) != 1 {
		t.Fatalf("expected 1 active silence, got %d", len(all))
	}
	if all[0].JobName != "job-a" {
		t.Errorf("expected job-a, got %s", all[0].JobName)
	}
}

func TestAll_Empty(t *testing.T) {
	r := silence.New()
	if got := r.All(time.Now()); len(got) != 0 {
		t.Fatalf("expected empty, got %d", len(got))
	}
}
