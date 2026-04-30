package webhook_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronscope/cronscope/internal/webhook"
)

func TestSend_Success(t *testing.T) {
	var received webhook.Payload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := webhook.NewClient(server.URL)
	p := webhook.Payload{
		JobName:   "backup",
		AlertType: webhook.AlertMissed,
		Message:   "job was not observed",
		Timestamp: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
	}

	if err := client.Send(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.JobName != p.JobName {
		t.Errorf("job name: got %q, want %q", received.JobName, p.JobName)
	}
	if received.AlertType != p.AlertType {
		t.Errorf("alert type: got %q, want %q", received.AlertType, p.AlertType)
	}
}

func TestSend_NonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := webhook.NewClient(server.URL)
	err := client.Send(webhook.Payload{JobName: "sync", AlertType: webhook.AlertLongRunning})
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestSend_InvalidURL(t *testing.T) {
	client := webhook.NewClient("http://127.0.0.1:0/no-listener")
	err := client.Send(webhook.Payload{JobName: "test", AlertType: webhook.AlertMissed})
	if err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}

func TestSend_TimestampAutoSet(t *testing.T) {
	var received webhook.Payload
	before := time.Now().UTC()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := webhook.NewClient(server.URL)
	// Send with zero timestamp — should be auto-filled.
	_ = client.Send(webhook.Payload{JobName: "nightly", AlertType: webhook.AlertMissed})

	if received.Timestamp.Before(before) {
		t.Errorf("expected auto-set timestamp >= %v, got %v", before, received.Timestamp)
	}
}
