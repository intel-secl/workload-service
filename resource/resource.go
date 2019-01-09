package resource

import (
	"fmt"
	"intel/isecl/workload-service/repository"
	"net/http"

	"github.com/jinzhu/gorm"

	"github.com/gorilla/mux"
)

// endpointSetter is a function that takes a Gorilla Mux Subrouter, and an instance of a WlsDatabase connection,
// and allows the end user to set and handle any API endpoints on that upaht
type endpointSetter func(r *mux.Router, db repository.WlsDatabase)

// endpointError is a custom error type that lets the thrower specify an http status code
type endpointError struct {
	Message    string
	StatusCode int
}

func (e endpointError) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}

// endpointHandler is the same as http.ResponseHandler, but returns an error that can be handled by a generic
// middleware handler
type endpointHandler func(w http.ResponseWriter, r *http.Request) error

func errorHandler(eh endpointHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := eh(w, r); err != nil {
			if gorm.IsRecordNotFoundError(err) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			switch t := err.(type) {
			case *endpointError:
				http.Error(w, t.Message, t.StatusCode)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}
