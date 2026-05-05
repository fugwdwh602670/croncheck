package notifier_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/user/croncheck/internal/notifier"
)

// TestSend_Unreachable verifies that a connection error is returned cleanly.
func TestSend_Unreachable(t *testing.T) {
	n := notifier.New("http://127.0.0.1:1") // nothing listening here
	err := n.Send(notifier.Alert{JobName: "job", AlertType: notifier.AlertFailed})
	if err == nil {
		t.Fatal("expected error for unreachable host, got nil")
	}
}

// TestSend_MultipleAlerts verifies sequential sends work correctly.
func TestSend_MultipleAlerts(t *testing.T) {
	var count int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := notifier.New(server.URL)
	alerts := []notifier.Alert{
		{JobName: "job-a", AlertType: notifier.AlertMissed, Message: "overdue"},
		{JobName: "job-b", AlertType: notifier.AlertFailed, Message: "exit code 1"},
		{JobName: "job-c", AlertType: notifier.AlertMissed, Message: "overdue"},
	}

	for _, a := range alerts {
		if err := n.Send(a); err != nil {
			t.Fatalf("send %q: %v", a.JobName, err)
		}
	}

	if got := atomic.LoadInt32(&count); got != int32(len(alerts)) {
		t.Errorf("expected %d requests, got %d", len(alerts), got)
	}
}
