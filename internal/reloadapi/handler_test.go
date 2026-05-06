package reloadapi_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cronscope/cronscope/internal/reloadapi"
)

type stubReloader struct {
	err error
}

func (s *stubReloader) Reload() error { return s.err }

func TestHandler_MethodNotAllowed(t *testing.T) {
	h := reloadapi.Handler(&stubReloader{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/reload", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_ReloadSuccess(t *testing.T) {
	h := reloadapi.Handler(&stubReloader{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/reload", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if ok, _ := body["ok"].(bool); !ok {
		t.Errorf("expected ok=true, got %v", body["ok"])
	}
	if msg, _ := body["message"].(string); !strings.Contains(msg, "reloaded") {
		t.Errorf("unexpected message: %q", msg)
	}
}

func TestHandler_ReloadError(t *testing.T) {
	h := reloadapi.Handler(&stubReloader{err: errors.New("disk read error")})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/reload", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if ok, _ := body["ok"].(bool); ok {
		t.Errorf("expected ok=false")
	}
	if msg, _ := body["message"].(string); !strings.Contains(msg, "disk read error") {
		t.Errorf("expected error message in response, got %q", msg)
	}
}

func TestHandler_ContentTypeJSON(t *testing.T) {
	h := reloadapi.Handler(&stubReloader{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/reload", nil))

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("expected application/json content-type, got %q", ct)
	}
}
