package retryapi_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/cronscope/internal/retryapi"
)

type fakeRetryer struct {
	suppressed map[string]bool
	retryErr   error
	retried    []string
}

func (f *fakeRetryer) Retry(job string) error {
	if f.retryErr != nil {
		return f.retryErr
	}
	f.retried = append(f.retried, job)
	return nil
}

func (f *fakeRetryer) IsSuppressed(job string) bool {
	return f.suppressed[job]
}

func newRetryer() *fakeRetryer {
	return &fakeRetryer{suppressed: map[string]bool{}}
}

func doRequest(t *testing.T, r retryapi.Retryer, method, url string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, url, nil)
	w := httptest.NewRecorder()
	retryapi.Handler(r).ServeHTTP(w, req)
	return w
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	w := doRequest(t, newRetryer(), http.MethodGet, "/retry?job=backup")
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandler_MissingJobParam(t *testing.T) {
	w := doRequest(t, newRetryer(), http.MethodPost, "/retry")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_SuppressedJob(t *testing.T) {
	r := newRetryer()
	r.suppressed["backup"] = true
	w := doRequest(t, r, http.MethodPost, "/retry?job=backup")
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestHandler_RetrySuccess(t *testing.T) {
	r := newRetryer()
	w := doRequest(t, r, http.MethodPost, "/retry?job=backup")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["retried"] != true {
		t.Errorf("expected retried=true, got %v", resp["retried"])
	}
	if resp["job"] != "backup" {
		t.Errorf("expected job=backup, got %v", resp["job"])
	}
	if len(r.retried) != 1 || r.retried[0] != "backup" {
		t.Errorf("expected retry to be called with 'backup', got %v", r.retried)
	}
}

func TestHandler_RetryError(t *testing.T) {
	r := newRetryer()
	r.retryErr = errors.New("webhook unreachable")
	w := doRequest(t, r, http.MethodPost, "/retry?job=backup")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	var resp map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp["message"] != "webhook unreachable" {
		t.Errorf("unexpected message: %v", resp["message"])
	}
}

func TestHandler_ContentTypeJSON(t *testing.T) {
	w := doRequest(t, newRetryer(), http.MethodPost, "/retry?job=db-snapshot")
	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}
