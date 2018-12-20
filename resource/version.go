package resource

import (
	"intel/isecl/workload-service/repository"
	"net/http"

	"github.com/gorilla/mux"
)

// SetVersionEndpoints installs route handler for GET /version
func SetVersionEndpoints(r *mux.Router, db repository.WlsDatabase) {
	r.HandleFunc("/version", getVersion)
}

// GetVersion handles GET /version
func getVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("1.0"))
}
