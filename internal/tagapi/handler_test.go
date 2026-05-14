package tagapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cronscope/internal/tagapi"
)

func doReq(t *testing.T, store *tagapi.Store, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var buf *bytes.Buffer
	if body != "" {
		buf = bytes.NewBufferString(body)
	} else {
		buf = &bytes.Buffer{}
	}
	req := httptest.NewRequest(method, path, buf)
	rec := httptest.NewRecorder()
	tagapi.Handler(store).ServeHTTP(rec, req)
	return rec
}

func TestHandler_GetMissingJobReturnsEmptyObject(t *testing.T) {
	store := tagapi.NewStore()
	rec := doReq(t, store, http.MethodGet, "/backup", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestHandler_PutAndGet(t *testing.T) {
	store := tagapi.NewStore()
	body := `{"env":"prod","team":"platform"}`
	rec := doReq(t, store, http.MethodPut, "/nightly-report", body)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	rec = doReq(t, store, http.MethodGet, "/nightly-report", "")
	var tags map[string]string
	json.NewDecoder(rec.Body).Decode(&tags)
	if tags["env"] != "prod" || tags["team"] != "platform" {
		t.Errorf("unexpected tags: %v", tags)
	}
}

func TestHandler_DeleteRemovesTags(t *testing.T) {
	store := tagapi.NewStore()
	store.Set("cleanup", map[string]string{"tier": "low"})

	rec := doReq(t, store, http.MethodDelete, "/cleanup", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	_, ok := store.Get("cleanup")
	if ok {
		t.Error("expected tags to be removed")
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	store := tagapi.NewStore()
	rec := doReq(t, store, http.MethodPost, "/somejob", "{}")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_MissingJobName(t *testing.T) {
	store := tagapi.NewStore()
	rec := doReq(t, store, http.MethodGet, "/", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandler_ContentTypeJSON(t *testing.T) {
	store := tagapi.NewStore()
	rec := doReq(t, store, http.MethodGet, "/anyjob", "")
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}
