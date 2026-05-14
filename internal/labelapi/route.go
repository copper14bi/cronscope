package labelapi

import "net/http"

// Register mounts the label handler under /labels/ on the given mux.
func Register(mux *http.ServeMux, s *Store) {
	mux.Handle("/labels/", http.StripPrefix("/labels", Handler(s)))
}
