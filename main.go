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
	defer db.Close()
	config.Postgres.Password = ""
	if err != nil {
		log.Fatal("could not open db", err)
	}
	r := mux.NewRouter().PathPrefix("/wls").Subrouter()
	// Set Resource Endpoints
	resource.SetFlavorsEndpoints(r.PathPrefix("/flavors").Subrouter(), db)
	// Set Image Endpoints
	resource.SetImagesEndpoints(r.PathPrefix("/images").Subrouter(), db)
	// Setup Version Endpoint
	resource.SetVersionEndpoints(r, db)
	if config.UseTLS {
		//http.ListenAndServe
	} else {
		http.ListenAndServe(fmt.Sprintf(":%d", config.Port), r)
	}
}
