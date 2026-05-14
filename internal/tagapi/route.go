package tagapi

import "net/http"

// Register mounts the tag handler under /tags/ on the given mux.
func Register(mux *http.ServeMux, store *Store) {
	mux.Handle("/tags/", http.StripPrefix("/tags", Handler(store)))
}
