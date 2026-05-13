package snoozeapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"cronscope/internal/snoozeapi"
)

type fakeSnoozer struct {
	mu      sync.Mutex
	records map[string]time.Time
}

func (f *fakeSnoozer) Snooze(job string, until time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.records[job] = until
}

func (f *fakeSnoozer) Unsnooze(job string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.records, job)
}

func (f *fakeSnoozer) SnoozedUntil(job string) (time.Time, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.records[job]
	return t, ok
}

func newFakeSnoozer() *fakeSnoozer {
	return &fakeSnoozer{records: make(map[string]time.Time)}
}

func doReq(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var buf *bytes.Buffer
	if body != "" {
		buf = bytes.NewBufferString(body)
	} else {
		buf = &bytes.Buffer{}
	}
	req := httptest.NewRequest(method, path, buf)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	h := snoozeapi.Handler(newFakeSnoozer())
	rec := doReq(t, h, http.MethodPatch, "/snooze/myjob", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_PostAndGet(t *testing.T) {
	s := newFakeSnoozer()
	h := snoozeapi.Handler(s)

	body := `{"duration":"30m"}`
	rec := doReq(t, h, http.MethodPost, "/snooze/backup", body)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	rec = doReq(t, h, http.MethodGet, "/snooze/backup", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if result["job"] != "backup" {
		t.Errorf("expected job=backup, got %v", result["job"])
	}
}

func TestHandler_DeleteRemovesSnooze(t *testing.T) {
	s := newFakeSnoozer()
	s.Snooze("reports", time.Now().Add(10*time.Minute))
	h := snoozeapi.Handler(s)

	rec := doReq(t, h, http.MethodDelete, "/snooze/reports", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	_, ok := s.SnoozedUntil("reports")
	if ok {
		t.Error("expected snooze to be removed")
	}
}

func TestHandler_GetNotFound(t *testing.T) {
	h := snoozeapi.Handler(newFakeSnoozer())
	rec := doReq(t, h, http.MethodGet, "/snooze/unknown", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_InvalidDuration(t *testing.T) {
	h := snoozeapi.Handler(newFakeSnoozer())
	rec := doReq(t, h, http.MethodPost, "/snooze/myjob", `{"duration":"notaduration"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
