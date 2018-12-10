package resource

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// SetVersionEndpoints installs route handler for GET /version
func SetVersionEndpoints(r *mux.Router, db *gorm.DB) {
	r.HandleFunc("/version", getVersion)
}

// GetVersion handles GET /version
func getVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("1.0"))
}
