package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/cronscope/internal/healthcheck"
)

// fakeState satisfies healthcheck.StateReader for testing.
type fakeState struct {
	names []string
}

func (f *fakeState) Names() []string { return f.names }

func TestHandler_StatusOK(t *testing.T) {
	st := &fakeState{names: []string{"backup", "report", "cleanup"}}
	h := healthcheck.Handler(st)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp healthcheck.Response
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status \"ok\", got %q", resp.Status)
	}
	if resp.JobCount != 3 {
		t.Errorf("expected job_count 3, got %d", resp.JobCount)
	}
	if resp.CheckedAt.IsZero() {
		t.Error("expected non-zero checked_at timestamp")
	}
}

func TestHandler_ContentTypeJSON(t *testing.T) {
	st := &fakeState{names: []string{"job1"}}
	h := healthcheck.Handler(st)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

func TestHandler_EmptyJobRegistry(t *testing.T) {
	st := &fakeState{names: []string{}}
	h := healthcheck.Handler(st)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp healthcheck.Response
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.JobCount != 0 {
		t.Errorf("expected job_count 0, got %d", resp.JobCount)
	}
}
