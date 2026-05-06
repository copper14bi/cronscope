package metricsapi

import (
	"net/http"

	"github.com/example/cronscope/internal/history"
	"github.com/example/cronscope/internal/state"
)

// Register mounts the metrics handler on the given ServeMux under the path
// /metrics. It is a convenience wrapper so callers do not need to import the
// package solely to call Handler directly.
//
//	// Example usage in main.go:
//	//   metricsapi.Register(mux, stateStore, historyStore)
func Register(mux *http.ServeMux, st *state.Store, hs *history.Store) {
	mux.Handle("/metrics", Handler(st, hs))
}
