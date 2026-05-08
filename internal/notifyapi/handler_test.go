package notifyapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cronscope/internal/notifyapi"
)

type fakeStore struct {
	records []notifyapi.NotifyRecord
}

func (f *fakeStore) Recent(limit int) []notifyapi.NotifyRecord {
	if len(f.records) <= limit {
		return f.records
	}
	return f.records[:limit]
}

func doGet(t *testing.T, store notifyapi.Store) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/notify", nil)
	notifyapi.Handler(store).ServeHTTP(rec, req)
	return rec
}

func TestHandler_ReturnsRecords(t *testing.T) {
	store := &fakeStore{
		records: []notifyapi.NotifyRecord{
			{JobName: "backup", Reason: "missed", SentAt: time.Now(), Cooldown: false},
			{JobName: "sync", Reason: "long_running", SentAt: time.Now(), Cooldown: true},
		},
	}

	rec := doGet(t, store)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var got []notifyapi.NotifyRecord
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 records, got %d", len(got))
	}
	if got[0].JobName != "backup" {
		t.Errorf("expected backup, got %s", got[0].JobName)
	}
}

func TestHandler_EmptyStore(t *testing.T) {
	store := &fakeStore{}
	rec := doGet(t, store)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var got []notifyapi.NotifyRecord
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d items", len(got))
	}
}

func TestHandler_ContentTypeJSON(t *testing.T) {
	rec := doGet(t, &fakeStore{})
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/notify", nil)
	notifyapi.Handler(&fakeStore{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
