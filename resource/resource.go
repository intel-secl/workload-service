package resource

import (
	"intel/isecl/workload-service/repository"
	"net/http"

	"github.com/gorilla/mux"
)

// EndpointSetter is a function that takes a Gorilla Mux Subrouter, and an instance of a WlsDatabase connection,
// and allows the end user to set and handle any API endpoints on that upaht
type EndpointSetter func(r *mux.Router, db repository.WlsDatabase)

// EndpointHandler is the same as http.ResponseHandler, but returns an error that can be haled by a generic
// middleware handelr
type EndpointHandler func(w http.ResponseWriter, r *http.Request) error
