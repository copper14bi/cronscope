package metricsapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/cronscope/internal/history"
	"github.com/example/cronscope/internal/metricsapi"
	"github.com/example/cronscope/internal/state"
)

func newStores(t *testing.T) (*state.Store, *history.Store) {
	t.Helper()
	return state.New(), history.New(50)
}

func seedHistory(hs *history.Store, job string, kind history.EventKind, n int) {
	for i := 0; i < n; i++ {
		hs.Record(history.Event{Job: job, Kind: kind, At: time.Now()})
	}
}

func TestMetrics_EmptyStore(t *testing.T) {
	st, hs := newStores(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsapi.Handler(st, hs).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp metricsapi.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.TotalJobs != 0 || resp.TotalAlerts != 0 {
		t.Errorf("expected zeroed counters, got jobs=%d alerts=%d", resp.TotalJobs, resp.TotalAlerts)
	}
}

func TestMetrics_CountsAlerts(t *testing.T) {
	st, hs := newStores(t)
	st.MarkSeen("backup", time.Now())
	st.MarkSeen("cleanup", time.Now())
	seedHistory(hs, "backup", history.KindMissed, 3)
	seedHistory(hs, "backup", history.KindLongRunning, 1)
	seedHistory(hs, "cleanup", history.KindMissed, 2)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsapi.Handler(st, hs).ServeHTTP(rec, req)

	var resp metricsapi.Response
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if resp.TotalJobs != 2 {
		t.Errorf("expected 2 jobs, got %d", resp.TotalJobs)
	}
	if resp.TotalAlerts != 6 {
		t.Errorf("expected 6 total alerts, got %d", resp.TotalAlerts)
	}
	if resp.Jobs["backup"].Missed != 3 {
		t.Errorf("backup missed: want 3, got %d", resp.Jobs["backup"].Missed)
	}
	if resp.Jobs["backup"].LongRunning != 1 {
		t.Errorf("backup long_running: want 1, got %d", resp.Jobs["backup"].LongRunning)
	}
	if resp.Jobs["cleanup"].Missed != 2 {
		t.Errorf("cleanup missed: want 2, got %d", resp.Jobs["cleanup"].Missed)
	}
}

func TestMetrics_ContentTypeJSON(t *testing.T) {
	st, hs := newStores(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsapi.Handler(st, hs).ServeHTTP(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestMetrics_MethodNotAllowed(t *testing.T) {
	st, hs := newStores(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	metricsapi.Handler(st, hs).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestMetrics_GeneratedAtPresent(t *testing.T) {
	st, hs := newStores(t)
	before := time.Now().UTC()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsapi.Handler(st, hs).ServeHTTP(rec, req)

	var resp metricsapi.Response
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.GeneratedAt.Before(before) {
		t.Errorf("generated_at %v is before request time %v", resp.GeneratedAt, before)
	}
}
