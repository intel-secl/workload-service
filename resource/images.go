package resource

import (
	"encoding/json"
	"fmt"
	"intel/isecl/lib/middleware/logger"
	"intel/isecl/workload-service/config"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// SetImagesEndpoints sets endpoints for /image
func SetImagesEndpoints(r *mux.Router, db repository.WlsDatabase) {
	logger := logger.NewLogger(config.LogWriter, "WLS - ", log.Ldate|log.Ltime)
	// r.HandleFunc("/{id}/flavors", nil).Methods("GET")
	// r.HandleFunc("/{id}/flavors/{flavorID}", nil).Methods("GET")
	// r.HandleFunc("/{id}/flavors/{flavorID}", nil).Methods("DELETE")
	r.HandleFunc("/{id}", logger(getImageByID(db))).Methods("GET")
	r.HandleFunc("/{id}", logger(deleteImageByID(db))).Methods("DELETE")
	r.HandleFunc("", logger(queryImages(db))).Methods("GET")
	r.HandleFunc("", logger(createImage(db))).Methods("POST").Headers("Content-Type", "application/json")
}

func getAssociatedFlavors(db repository.WlsDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func queryImages(db repository.WlsDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locator := repository.ImageLocator{}
		flavorID, ok := r.URL.Query()["flavor_id"]
		if ok && len(flavorID) >= 1 {
			locator.FlavorID = flavorID[0]
		}

		images, err := db.ImageRepository().RetrieveByFilterCriteria(locator)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(images) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := json.NewEncoder(w).Encode(images); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
		}
	}
}

func getImageByID(db repository.WlsDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := mux.Vars(r)["id"]
		image, err := db.ImageRepository().RetrieveByUUID(uuid)
		if err == nil {
			mErr := json.NewEncoder(w).Encode(image)
			if mErr != nil {
				http.Error(w, mErr.Error(), http.StatusInternalServerError)
			}
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
		} else if gorm.IsRecordNotFoundError(err) {
			var code int
			if gorm.IsRecordNotFoundError(err) {
				code = http.StatusNotFound
			} else {
				code = http.StatusInternalServerError
			}
			http.Error(w, err.Error(), code)
		}
	}
}

func deleteImageByID(db repository.WlsDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := mux.Vars(r)["id"]
		err := db.ImageRepository().DeleteByUUID(uuid)
		if err != nil {
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

func createImage(db repository.WlsDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var formBody model.Image
		if err := json.NewDecoder(r.Body).Decode(&formBody); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := db.ImageRepository().Create(&formBody); err != nil {
			switch err {
			case repository.ErrImageAssociationAlreadyExists:
				http.Error(w, fmt.Sprintf("image with UUID %s is already associated with a flavor", formBody.ID), http.StatusConflict)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		w.WriteHeader(http.StatusCreated)
	}
}
