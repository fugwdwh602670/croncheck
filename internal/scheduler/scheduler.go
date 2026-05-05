package scheduler

import (
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"croncheck/internal/config"
)

// JobStatus tracks the last execution state of a cron job.
type JobStatus struct {
	Name      string
	Schedule  string
	LastSeen  time.Time
	Missed    bool
	mu        sync.RWMutex
}

// Scheduler manages job heartbeat tracking.
type Scheduler struct {
	jobs    map[string]*JobStatus
	parser  cron.Parser
	mu      sync.RWMutex
}

// New creates a Scheduler pre-populated from config.
func New(jobs []config.Job) *Scheduler {
	s := &Scheduler{
		jobs:   make(map[string]*JobStatus, len(jobs)),
		parser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
	for _, j := range jobs {
		s.jobs[j.Name] = &JobStatus{
			Name:     j.Name,
			Schedule: j.Schedule,
		}
	}
	return s
}

// Heartbeat records a successful execution for the named job.
func (s *Scheduler) Heartbeat(name string) bool {
	s.mu.RLock()
	job, ok := s.jobs[name]
	s.mu.RUnlock()
	if !ok {
		return false
	}
	job.mu.Lock()
	job.LastSeen = time.Now()
	job.Missed = false
	job.mu.Unlock()
	log.Printf("[heartbeat] job=%s time=%s", name, job.LastSeen.Format(time.RFC3339))
	return true
}

// CheckMissed evaluates all jobs and returns those that have missed their schedule.
func (s *Scheduler) CheckMissed(now time.Time) []*JobStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var missed []*JobStatus
	for _, job := range s.jobs {
		if isMissed(job, s.parser, now) {
			job.mu.Lock()
			job.Missed = true
			job.mu.Unlock()
			missed = append(missed, job)
		}
	}
	return missed
}

func isMissed(job *JobStatus, parser cron.Parser, now time.Time) bool {
	sched, err := parser.Parse(job.Schedule)
	if err != nil {
		log.Printf("[warn] invalid schedule for job=%s: %v", job.Name, err)
		return false
	}
	job.mu.RLock()
	lastSeen := job.LastSeen
	job.mu.RUnlock()
	if lastSeen.IsZero() {
		return false // not yet tracked
	}
	nextExpected := sched.Next(lastSeen)
	gracePeriod := 2 * time.Minute
	return now.After(nextExpected.Add(gracePeriod))
}
