package commentapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronscope/cronscope/internal/commentapi"
)

func newStore() *commentapi.Store { return commentapi.NewStore() }

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

func TestHandler_GetMissingJobReturnsEmptyComment(t *testing.T) {
	h := commentapi.Handler(newStore())
	rec := doReq(t, h, http.MethodGet, "/backup", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["comment"] != "" {
		t.Errorf("expected empty comment, got %q", resp["comment"])
	}
}

func TestHandler_PutAndGet(t *testing.T) {
	store := newStore()
	h := commentapi.Handler(store)

	rec := doReq(t, h, http.MethodPut, "/nightly", `{"comment":"disabled for deploy"}`)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT: expected 204, got %d", rec.Code)
	}

	rec = doReq(t, h, http.MethodGet, "/nightly", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET: expected 200, got %d", rec.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["comment"] != "disabled for deploy" {
		t.Errorf("unexpected comment: %q", resp["comment"])
	}
}

func TestHandler_DeleteRemovesComment(t *testing.T) {
	store := newStore()
	store.Set("cleanup", "some note")
	h := commentapi.Handler(store)

	rec := doReq(t, h, http.MethodDelete, "/cleanup", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE: expected 204, got %d", rec.Code)
	}

	if c, ok := store.Get("cleanup"); ok || c != "" {
		t.Errorf("comment should be deleted, got %q ok=%v", c, ok)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	h := commentapi.Handler(newStore())
	rec := doReq(t, h, http.MethodPost, "/job1", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_ContentTypeJSON(t *testing.T) {
	h := commentapi.Handler(newStore())
	rec := doReq(t, h, http.MethodGet, "/anyjob", "")
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestHandler_PutInvalidJSON(t *testing.T) {
	h := commentapi.Handler(newStore())
	rec := doReq(t, h, http.MethodPut, "/job1", `not-json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
