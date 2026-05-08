package retryapi

import "net/http"

// Register mounts the retry handler under POST /retry on the given mux.
func Register(mux *http.ServeMux, r Retryer) {
	mux.Handle("/retry", Handler(r))
}
