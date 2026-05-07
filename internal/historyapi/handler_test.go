package historyapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/cronscope/internal/history"
	"github.com/yourorg/cronscope/internal/historyapi"
)

func seedStore() *history.Store {
	s := history.New(50)
	s.Record(history.Event{JobName: "backup", Type: history.EventMissed, OccurredAt: time.Now().UTC()})
	s.Record(history.Event{JobName: "sync", Type: history.EventLongRunning, OccurredAt: time.Now().UTC()})
	s.Record(history.Event{JobName: "backup", Type: history.EventRecovered, OccurredAt: time.Now().UTC()})
	return s
}

// serveAndDecode is a helper that fires a GET request against the handler and
// decodes the JSON response body into dest. It returns the recorded response
// so callers can inspect status codes and headers.
func serveAndDecode(t *testing.T, s *history.Store, path string, dest interface{}) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	historyapi.Handler(s).ServeHTTP(rec, req)
	if err := json.NewDecoder(rec.Body).Decode(dest); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	return rec
}

func TestHandler_AllEvents(t *testing.T) {
	s := seedStore()
	var events []history.Event
	rec := serveAndDecode(t, s, "/history", &events)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}
}

func TestHandler_FilterByJob(t *testing.T) {
	s := seedStore()
	var events []history.Event
	rec := serveAndDecode(t, s, "/history/backup", &events)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events for 'backup', got %d", len(events))
	}
}

func TestHandler_ContentTypeJSON(t *testing.T) {
	s := history.New(10)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	historyapi.Handler(s).ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	s := history.New(10)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/history", nil)
	historyapi.Handler(s).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_EmptyStoreReturnsArray(t *testing.T) {
	s := history.New(10)
	var events []history.Event
	serveAndDecode(t, s, "/history", &events)

	if events == nil {
		t.Error("expected empty array, not null")
	}
}
