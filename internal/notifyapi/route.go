package notifyapi

import "net/http"

// Register mounts the notify API handler under /notify on the given mux.
func Register(mux *http.ServeMux, store Store) {
	mux.Handle("/notify", Handler(store))
}
