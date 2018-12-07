package resources

import "net/http"

// GetVersion handles GET /version
func GetVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("1.0"))
}
