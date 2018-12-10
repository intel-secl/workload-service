package main

import (
	"fmt"
	"intel/isecl/workload-service/resource"
	"log"
	"net/http"

	"github.com/jinzhu/gorm"

	"intel/isecl/workload-service/config"

	"github.com/gorilla/mux"
	// Import Postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	// source configuration from somewhere

	var sslMode string
	if config.Postgres.SSLMode {
		sslMode = "enable"
	} else {
		sslMode = "disable"
	}
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		config.Postgres.Hostname, config.Postgres.Port, config.Postgres.User, config.Postgres.DBName, config.Postgres.Password, sslMode))
	config.Postgres.Password = ""
	if err != nil {
		log.Fatal("could not open db", err)
	}
	r := mux.NewRouter()
	// Set Resource Endpoints
	resource.SetFlavorEndpoints(r, db)
	// Setup Version Endpoint
	resource.SetVersionEndpoints(r, db)
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// default handler
	})
	if config.UseTLS {
		//http.ListenAndServe
	} else {
		http.ListenAndServe(fmt.Sprintf(":%d", config.Port), r)
	}
}
