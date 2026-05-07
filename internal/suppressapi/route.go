package suppressapi

import (
	"net/http"

	"github.com/cronscope/cronscope/internal/alertmanager"
)

// Register mounts the suppression handler at /suppress on the given mux.
func Register(mux *http.ServeMux, am *alertmanager.Manager) {
	mux.Handle("/suppress", Handler(am))
}
