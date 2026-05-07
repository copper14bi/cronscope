package suppressapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronscope/cronscope/internal/alertmanager"
	"github.com/cronscope/cronscope/internal/suppressapi"
	"github.com/cronscope/cronscope/internal/webhook"
)

func newManager(t *testing.T) *alertmanager.Manager {
	t.Helper()
	client := webhook.NewClient("http://localhost", 5*time.Second)
	return alertmanager.New(client, time.Minute)
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	h := suppressapi.Handler(newManager(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/suppress", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandler_PostAndGet(t *testing.T) {
	am := newManager(t)
	h := suppressapi.Handler(am)

	body := `{"job":"backup","duration":"2h"}`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/suppress", bytes.NewBufferString(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/suppress", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var rules []alertmanager.SuppressionRule
	if err := json.NewDecoder(rec2.Body).Decode(&rules); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(rules) != 1 || rules[0].Job != "backup" {
		t.Fatalf("expected backup suppression, got %+v", rules)
	}
}

func TestHandler_DeleteRemovesSuppression(t *testing.T) {
	am := newManager(t)
	h := suppressapi.Handler(am)

	am.Suppress("cleanup", 2*time.Hour)

	body := `{"job":"cleanup"}`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/suppress", bytes.NewBufferString(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if len(am.Suppressions()) != 0 {
		t.Fatal("expected no active suppressions after delete")
	}
}

func TestHandler_PostInvalidDuration(t *testing.T) {
	h := suppressapi.Handler(newManager(t))
	body := `{"job":"backup","duration":"notaduration"}`
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/suppress", bytes.NewBufferString(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandler_ContentTypeJSON(t *testing.T) {
	h := suppressapi.Handler(newManager(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/suppress", nil))
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected application/json, got %q", ct)
	}
}
