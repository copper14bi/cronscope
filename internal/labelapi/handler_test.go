package labelapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func doReq(t *testing.T, h http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHandler_PutAndGet(t *testing.T) {
	s := NewStore()
	h := Handler(s)

	body, _ := json.Marshal(map[string]string{"env": "prod", "team": "ops"})
	rec := doReq(t, h, http.MethodPut, "/backup", body)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT: want 204, got %d", rec.Code)
	}

	rec = doReq(t, h, http.MethodGet, "/backup", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET: want 200, got %d", rec.Code)
	}
	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got["env"] != "prod" || got["team"] != "ops" {
		t.Errorf("unexpected labels: %v", got)
	}
}

func TestHandler_GetMissingJobReturnsEmptyObject(t *testing.T) {
	s := NewStore()
	h := Handler(s)

	rec := doReq(t, h, http.MethodGet, "/nonexistent", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty map, got %v", got)
	}
}

func TestHandler_DeleteRemovesLabels(t *testing.T) {
	s := NewStore()
	s.Set("cleanup", map[string]string{"tier": "low"})
	h := Handler(s)

	rec := doReq(t, h, http.MethodDelete, "/cleanup", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE: want 204, got %d", rec.Code)
	}
	if labels := s.Get("cleanup"); len(labels) != 0 {
		t.Errorf("expected labels deleted, got %v", labels)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	s := NewStore()
	h := Handler(s)

	rec := doReq(t, h, http.MethodPatch, "/somejob", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rec.Code)
	}
}

func TestHandler_ContentTypeJSON(t *testing.T) {
	s := NewStore()
	h := Handler(s)

	rec := doReq(t, h, http.MethodGet, "/anyjob", nil)
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("want application/json, got %s", ct)
	}
}

func TestHandler_InvalidJSON(t *testing.T) {
	s := NewStore()
	h := Handler(s)

	rec := doReq(t, h, http.MethodPut, "/job", []byte("not-json"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}
