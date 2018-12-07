package main

import (
	"fmt"
	"intel/isecl/workload-service/resource"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/version", resource.GetVersion)
	http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), r)
}
