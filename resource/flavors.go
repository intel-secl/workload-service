package resource

import (
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/config"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"intel/isecl/lib/middleware/logger"
	"intel/isecl/workload-service/repository"

	"github.com/gorilla/mux"
)

// SetFlavorsEndpoints
func SetFlavorsEndpoints(r *mux.Router, db repository.WlsDatabase) {
	logger := logger.NewLogger(config.LogWriter, "WLS - ", log.Ldate|log.Ltime)
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", logger(errorHandler(getFlavorByID(db)))).Methods("GET")
	r.HandleFunc("/{label}", logger(errorHandler(getFlavorByLabel(db)))).Methods("GET")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", logger(errorHandler(deleteFlavorByID(db)))).Methods("DELETE")
	r.HandleFunc("", logger(errorHandler(createFlavor(db)))).Methods("POST").Headers("Content-Type", "application/json")
}

func getFlavorByID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		fr := db.FlavorRepository()
		flavor, err := fr.RetrieveByUUID(id)
		if err != nil {
			return err
		} 
		if err := json.NewEncoder(w).Encode(flavor); err != nil {
			return err
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		return nil
	}
}

func getFlavorByLabel(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		label := mux.Vars(r)["label"]
		flavor, err := db.FlavorRepository().RetrieveByLabel(label)
		if err != nil {
			return err
		} 
		if err := json.NewEncoder(w).Encode(flavor); err != nil {
			return err
		} 
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		return nil
	}
}

func deleteFlavorByID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		fr := db.FlavorRepository()
		if err := fr.DeleteByUUID(id); err != nil {
			return err
		} 
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}

func createFlavor(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var f model.Flavor
		if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		// it's almost silly that we unmarshal, then remarshal it to store it back into the database, but at least it provides some validation of the input
		fr := db.FlavorRepository()

		// Performance Related:
		// currently, we don't decipher the creation error to see if Creation failed because a collision happened between a primary or unique key.
		// It would be nice to know why the record creation fails, and return the proper http status code.
		// It could be done several ways:
		// - Type assert the error back to PSQL (should be done in the repository layer), and bubble up that information somehow
		// - Manually run a query to see if anything exists with uuid or label (should be done in the repository layer, so we can execute it in a transaction)
		//    - Currently doing this ^
		switch err := fr.Create(&f); err {
		case repository.ErrFlavorLabelAlreadyExists:
			return &endpointError {
				Message: fmt.Sprintf("Flavor with Label %s already exists", f.Image.Meta.Description.Label), 
				StatusCode: http.StatusConflict,
			}
		case repository.ErrFlavorUUIDAlreadyExists:
			return &endpointError {
				Message: fmt.Sprintf("Flavor with UUID %s already exists", f.Image.Meta.ID), 
				StatusCode: http.StatusConflict,
			}
		case nil:
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(f); err != nil {
				return err
			}
			return nil
		default:
			return err 
		}
	}
}
