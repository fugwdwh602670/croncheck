package retention_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/croncheck/internal/retention"
)

type mockPruner struct {
	mu      sync.Mutex
	calls   []time.Duration
	returns []string
}

func (m *mockPruner) PruneStale(olderThan time.Duration) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, olderThan)
	return m.returns
}

func (m *mockPruner) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

func TestRun_CallsPrunerOnInterval(t *testing.T) {
	p := &mockPruner{returns: []string{}}
	c := retention.New(p, 10*time.Minute, 20*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Millisecond)
	defer cancel()

	c.Run(ctx)

	if p.callCount() < 2 {
		t.Errorf("expected at least 2 prune calls, got %d", p.callCount())
	}
}

func TestRun_PassesTTL(t *testing.T) {
	ttl := 30 * time.Minute
	p := &mockPruner{returns: []string{}}
	c := retention.New(p, ttl, 20*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Millisecond)
	defer cancel()

	c.Run(ctx)

	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.calls) == 0 {
		t.Fatal("expected at least one prune call")
	}
	if p.calls[0] != ttl {
		t.Errorf("expected TTL %v, got %v", ttl, p.calls[0])
	}
}

func TestRun_StopsOnContextCancel(t *testing.T) {
	p := &mockPruner{returns: []string{}}
	c := retention.New(p, 5*time.Minute, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		c.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Error("Run did not stop after context cancellation")
	}
}
