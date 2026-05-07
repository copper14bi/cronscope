package pauseapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/example/cronscope/internal/pauseapi"
)

// fakePauser is an in-memory Pauser for testing.
type fakePauser struct {
	mu     sync.Mutex
	paused map[string]time.Time
}

func newFakePauser() *fakePauser {
	return &fakePauser{paused: make(map[string]time.Time)}
}

func (f *fakePauser) Pause(job string, until time.Time) {
	f.mu.Lock(); defer f.mu.Unlock()
	f.paused[job] = until
}
func (f *fakePauser) Resume(job string) {
	f.mu.Lock(); defer f.mu.Unlock()
	delete(f.paused, job)
}
func (f *fakePauser) IsPaused(job string) bool {
	_, ok := f.PausedUntil(job)
	return ok
}
func (f *fakePauser) PausedUntil(job string) (time.Time, bool) {
	f.mu.Lock(); defer f.mu.Unlock()
	t, ok := f.paused[job]
	return t, ok
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	h := pauseapi.Handler(newFakePauser())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/backup", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_MissingJobName(t *testing.T) {
	h := pauseapi.Handler(newFakePauser())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandler_PostPausesJob(t *testing.T) {
	p := newFakePauser()
	h := pauseapi.Handler(p)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/backup?duration=2h", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !p.IsPaused("backup") {
		t.Fatal("expected job to be paused")
	}
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["paused"] != true {
		t.Errorf("expected paused=true in response")
	}
}

func TestHandler_DeleteResumesJob(t *testing.T) {
	p := newFakePauser()
	p.Pause("backup", time.Now().Add(time.Hour))
	h := pauseapi.Handler(p)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/backup", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if p.IsPaused("backup") {
		t.Fatal("expected job to be resumed")
	}
}

func TestHandler_GetReturnsStatus(t *testing.T) {
	p := newFakePauser()
	p.Pause("nightly", time.Now().Add(30*time.Minute))
	h := pauseapi.Handler(p)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nightly", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["paused"] != true {
		t.Errorf("expected paused=true")
	}
	if resp["until"] == "" {
		t.Errorf("expected until field to be set")
	}
}

func TestHandler_PostInvalidDuration(t *testing.T) {
	h := pauseapi.Handler(newFakePauser())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/backup?duration=notaduration", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
