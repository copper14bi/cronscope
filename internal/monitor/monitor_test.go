package monitor_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronscope/cronscope/internal/config"
	"github.com/cronscope/cronscope/internal/monitor"
	"github.com/cronscope/cronscope/internal/webhook"
)

func newTestSetup(t *testing.T) (*monitor.Monitor, *[]webhook.Payload, *httptest.Server) {
	t.Helper()
	var received []webhook.Payload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p webhook.Payload
		_ = json.NewDecoder(r.Body).Decode(&p)
		received = append(received, p)
		w.WriteHeader(http.StatusOK)
	}))
	cfg := &config.Config{
		WebhookURL: server.URL,
		Jobs: []config.Job{
			{Name: "backup", Schedule: "0 2 * * *", MaxDuration: 30 * time.Minute},
		},
	}
	client := webhook.NewClient(server.URL)
	m := monitor.New(cfg, client)
	return m, &received, server
}

func TestCheck_MissedJob(t *testing.T) {
	m, received, server := newTestSetup(t)
	defer server.Close()

	// Simulate no last-seen; check well after expected run
	now := time.Date(2024, 1, 10, 3, 0, 0, 0, time.UTC)
	m.Check(now)

	if len(*received) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(*received))
	}
	if (*received)[0].Event != "missed" {
		t.Errorf("expected event 'missed', got %q", (*received)[0].Event)
	}
}

func TestCheck_LongRunningJob(t *testing.T) {
	m, received, server := newTestSetup(t)
	defer server.Close()

	start := time.Date(2024, 1, 10, 2, 0, 0, 0, time.UTC)
	m.RecordStart("backup", start)

	// Check after max duration exceeded
	now := start.Add(45 * time.Minute)
	m.Check(now)

	if len(*received) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(*received))
	}
	if (*received)[0].Event != "long_running" {
		t.Errorf("expected event 'long_running', got %q", (*received)[0].Event)
	}
}

func TestCheck_HealthyJob(t *testing.T) {
	m, received, server := newTestSetup(t)
	defer server.Close()

	start := time.Date(2024, 1, 10, 2, 0, 0, 0, time.UTC)
	m.RecordStart("backup", start)
	m.RecordFinish("backup", start.Add(10*time.Minute))

	// Check shortly after finish — no alerts expected
	now := start.Add(12 * time.Minute)
	m.Check(now)

	if len(*received) != 0 {
		t.Errorf("expected no alerts, got %d", len(*received))
	}
}
