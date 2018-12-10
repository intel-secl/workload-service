package resource

import (
	"encoding/json"
	"fmt"
	"net/http"

	"intel/isecl/lib/flavor"
	"intel/isecl/workload-service/repository"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// SetFlavorEndpoints
func SetFlavorEndpoints(r *mux.Router, db *gorm.DB) {
	r.HandleFunc("/flavors/{id}", getFlavorByID(db)).Methods("GET")
	r.HandleFunc("/flavors/{id}", deleteFlavorByID(db)).Methods("DELETE")
	r.HandleFunc("/flavors", createFlavor(db)).Methods("POST").Headers("Content-Type", "application/json")
}

func getFlavorByID(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		fr := repository.GetFlavorRepository(db)
		flavor, err := fr.RetrieveByUUID(id)
		if err != nil {
			var code int
			if gorm.IsRecordNotFoundError(err) {
				code = http.StatusNotFound
			} else {
				code = http.StatusInternalServerError
			}
			http.Error(w, err.Error(), code)
		} else {
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(flavor)
			}
		}
	}
}

func deleteFlavorByID(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		fr := repository.GetFlavorRepository(db)
		if err := fr.DeleteByUUID(id); err != nil {
			var code int
			if gorm.IsRecordNotFoundError(err) {
				code = http.StatusNotFound
			} else {
				code = http.StatusInternalServerError
			}
			http.Error(w, err.Error(), code)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func createFlavor(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var f flavor.ImageFlavor
		if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		// it's almost silly that we unmarshal, then remarshal it to store it back into the database, but at least it provides some validation of the input
		fr := repository.GetFlavorRepository(db)

		// Performance Related:
		// currently, we don't decipher the creation error to see if Creation failed because a collision happened between a primary or unique key.
		// It would be nice to know why the record creation fails, and return the proper http status code.
		// It could be done several ways:
		// - Type assert the error back to PSQL (should be done in the repository layer), and bubble up that information somehow
		// - Manually run a query to see if anything exists with uuid or label (should be done in the repository layer, so we can execute it in a transaction)
		//    - Currently doing this ^
		switch err := fr.Create(&f); err {
		case repository.ErrFlavorLabelAlreadyExists:
			http.Error(w, fmt.Sprintf("Flavor with Label %s already exists", f.Image.Meta.Description.Label), http.StatusConflict)
		case repository.ErrFlavorUUIDAlreadyExists:
			http.Error(w, fmt.Sprintf("Flavor with UUID %s already exists", f.Image.Meta.ID), http.StatusConflict)
		case nil:
			w.WriteHeader(http.StatusCreated)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
