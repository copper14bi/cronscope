package silenceapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/cronscope/cronscope/internal/silenceapi"
)

type fakeSilencer struct {
	mu       sync.Mutex
	silences map[string]time.Time
}

func newFakeSilencer() *fakeSilencer {
	return &fakeSilencer{silences: make(map[string]time.Time)}
}

func (f *fakeSilencer) Silence(job string, until time.Time) {
	f.mu.Lock(); defer f.mu.Unlock()
	f.silences[job] = until
}

func (f *fakeSilencer) Unsilence(job string) {
	f.mu.Lock(); defer f.mu.Unlock()
	delete(f.silences, job)
}

func (f *fakeSilencer) IsSilenced(job string) (bool, time.Time) {
	f.mu.Lock(); defer f.mu.Unlock()
	until, ok := f.silences[job]
	return ok, until
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
	h := silenceapi.Handler(newFakeSilencer())
	rec := doReq(t, h, http.MethodPatch, "/backups", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_PostAndGet(t *testing.T) {
	s := newFakeSilencer()
	h := silenceapi.Handler(s)

	rec := doReq(t, h, http.MethodPost, "/backups", `{"duration":"2h"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	rec = doReq(t, h, http.MethodGet, "/backups", "")
	var resp struct {
		Silenced bool      `json:"silenced"`
		Until    time.Time `json:"until"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Silenced {
		t.Error("expected job to be silenced")
	}
	if time.Until(resp.Until) < time.Hour {
		t.Error("expected silence to last at least 1h")
	}
}

func TestHandler_DeleteRemovesSilence(t *testing.T) {
	s := newFakeSilencer()
	s.Silence("nightly", time.Now().Add(time.Hour))
	h := silenceapi.Handler(s)

	rec := doReq(t, h, http.MethodDelete, "/nightly", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	silenced, _ := s.IsSilenced("nightly")
	if silenced {
		t.Error("expected silence to be removed")
	}
}

func TestHandler_PostInvalidDuration(t *testing.T) {
	h := silenceapi.Handler(newFakeSilencer())
	rec := doReq(t, h, http.MethodPost, "/job1", `{"duration":"bad"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandler_ContentTypeJSON(t *testing.T) {
	h := silenceapi.Handler(newFakeSilencer())
	rec := doReq(t, h, http.MethodGet, "/anyjob", "")
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}
