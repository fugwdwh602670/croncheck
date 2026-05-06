package ratelimit_test

import (
	"testing"
	"time"

	"github.com/croncheck/internal/ratelimit"
)

func TestAllow_FirstAlert(t *testing.T) {
	l := ratelimit.New(1 * time.Minute)
	if !l.Allow("backup") {
		t.Fatal("expected first alert to be allowed")
	}
}

func TestAllow_WithinCooldown(t *testing.T) {
	l := ratelimit.New(1 * time.Minute)
	l.Allow("backup") // record first
	if l.Allow("backup") {
		t.Fatal("expected second alert within cooldown to be suppressed")
	}
}

func TestAllow_AfterCooldown(t *testing.T) {
	l := ratelimit.New(10 * time.Millisecond)
	l.Allow("backup")
	time.Sleep(20 * time.Millisecond)
	if !l.Allow("backup") {
		t.Fatal("expected alert to be allowed after cooldown elapsed")
	}
}

func TestAllow_IndependentJobs(t *testing.T) {
	l := ratelimit.New(1 * time.Minute)
	l.Allow("job-a")
	if !l.Allow("job-b") {
		t.Fatal("expected independent job to be allowed")
	}
}

func TestReset_AllowsImmediately(t *testing.T) {
	l := ratelimit.New(1 * time.Minute)
	l.Allow("backup")
	l.Reset("backup")
	if !l.Allow("backup") {
		t.Fatal("expected alert to be allowed after reset")
	}
}

func TestLastAlert_UnknownJob(t *testing.T) {
	l := ratelimit.New(1 * time.Minute)
	_, ok := l.LastAlert("nonexistent")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestLastAlert_KnownJob(t *testing.T) {
	l := ratelimit.New(1 * time.Minute)
	before := time.Now()
	l.Allow("backup")
	t2, ok := l.LastAlert("backup")
	if !ok {
		t.Fatal("expected entry for known job")
	}
	if t2.Before(before) {
		t.Errorf("last alert time %v is before test start %v", t2, before)
	}
}
