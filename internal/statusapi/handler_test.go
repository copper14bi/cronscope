package statusapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronscope/cronscope/internal/state"
	"github.com/cronscope/cronscope/internal/statusapi"
)

// fakeStore satisfies statusapi.Store for testing.
type fakeStore struct {
	entries map[string]state.Entry
}

func (f *fakeStore) Names() []string {
	names := make([]string, 0, len(f.entries))
	for k := range f.entries {
		names = append(names, k)
	}
	return names
}

func (f *fakeStore) Get(name string) (state.Entry, bool) {
	e, ok := f.entries[name]
	return e, ok
}

func newStore(entries map[string]state.Entry) *fakeStore {
	return &fakeStore{entries: entries}
}

func TestHandler_ReturnsAllJobs(t *testing.T) {
	now := time.Now().UTC()
	store := newStore(map[string]state.Entry{
		"backup": {LastSeen: now, Running: false},
		"sync":   {LastSeen: now, Running: true, RunSince: now.Add(-5 * time.Minute)},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	statusapi.Handler(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp statusapi.StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp.Jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(resp.Jobs))
	}
}

func TestHandler_RunningJobIncludesRunSince(t *testing.T) {
	now := time.Now().UTC()
	store := newStore(map[string]state.Entry{
		"etl": {Running: true, RunSince: now.Add(-10 * time.Minute)},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	statusapi.Handler(store).ServeHTTP(rec, req)

	var resp statusapi.StatusResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp.Jobs) != 1 {
		t.Fatalf("expected 1 job")
	}
	if !resp.Jobs[0].Running {
		t.Error("expected running=true")
	}
	if resp.Jobs[0].RunSince == nil {
		t.Error("expected run_since to be set for running job")
	}
}

func TestHandler_ContentTypeJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	statusapi.Handler(newStore(nil)).ServeHTTP(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("unexpected Content-Type: %s", ct)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/status", nil)
	statusapi.Handler(newStore(nil)).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
