package main

import (
	"fmt"
	"net/http"

	"intel/isecl/workload-service/config"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	// Setup Version Endpoint
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// default handler
	})
	if config.UseTLS {
		//http.ListenAndServe
	} else {
		http.ListenAndServe(fmt.Sprintf(":%d", config.Port), r)
	}
}
