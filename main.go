package main

import (
	"fmt"
	"intel/isecl/workload-service/routes"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/version", routes.GetVersion)
	http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), r)
}
