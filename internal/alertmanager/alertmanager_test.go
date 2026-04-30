package alertmanager_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/cronscope/internal/alertmanager"
	"github.com/yourorg/cronscope/internal/webhook"
)

func newTestManager(t *testing.T, cooldown time.Duration) (*alertmanager.Manager, *int32) {
	t.Helper()
	var count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	client := webhook.NewClient(server.URL)
	return alertmanager.New(client, cooldown), &count
}

func makeAlert(job string, typ alertmanager.AlertType) alertmanager.Alert {
	return alertmanager.Alert{
		JobName:   job,
		Type:      typ,
		Message:   "test alert",
		Timestamp: time.Now(),
	}
}

func TestSend_DispatchesAlert(t *testing.T) {
	mgr, count := newTestManager(t, time.Minute)
	sent, err := mgr.Send(makeAlert("backup", alertmanager.AlertMissed))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sent {
		t.Fatal("expected alert to be sent")
	}
	if atomic.LoadInt32(count) != 1 {
		t.Fatalf("expected 1 webhook call, got %d", atomic.LoadInt32(count))
	}
}

func TestSend_DeduplicatesWithinCooldown(t *testing.T) {
	mgr, count := newTestManager(t, time.Minute)
	a := makeAlert("backup", alertmanager.AlertMissed)
	mgr.Send(a) //nolint:errcheck
	sent, err := mgr.Send(a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sent {
		t.Fatal("expected second alert to be suppressed")
	}
	if atomic.LoadInt32(count) != 1 {
		t.Fatalf("expected 1 webhook call, got %d", atomic.LoadInt32(count))
	}
}

func TestSend_AllowsAfterCooldown(t *testing.T) {
	mgr, count := newTestManager(t, time.Millisecond)
	a := makeAlert("backup", alertmanager.AlertMissed)
	mgr.Send(a) //nolint:errcheck
	time.Sleep(5 * time.Millisecond)
	sent, err := mgr.Send(a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sent {
		t.Fatal("expected alert to be sent after cooldown")
	}
	if atomic.LoadInt32(count) != 2 {
		t.Fatalf("expected 2 webhook calls, got %d", atomic.LoadInt32(count))
	}
}

func TestReset_ClearsCooldown(t *testing.T) {
	mgr, count := newTestManager(t, time.Minute)
	a := makeAlert("backup", alertmanager.AlertMissed)
	mgr.Send(a) //nolint:errcheck
	mgr.Reset("backup")
	sent, err := mgr.Send(a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sent {
		t.Fatal("expected alert after reset")
	}
	if atomic.LoadInt32(count) != 2 {
		t.Fatalf("expected 2 webhook calls, got %d", atomic.LoadInt32(count))
	}
}

func TestSend_AlertPayloadIsJSON(t *testing.T) {
	var received alertmanager.Alert
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	client := webhook.NewClient(server.URL)
	mgr := alertmanager.New(client, time.Minute)
	a := makeAlert("cleanup", alertmanager.AlertLongRunning)
	mgr.Send(a) //nolint:errcheck
	if received.JobName != "cleanup" {
		t.Errorf("expected job_name=cleanup, got %q", received.JobName)
	}
}
