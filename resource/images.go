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
func SetImagesEndpoints(r *mux.Router, db *gorm.DB) {
	logger := logger.NewLogger(config.LogWriter, "WLS - ", log.Ldate|log.Ltime)
	r.HandleFunc("/{id}", logger(getImageByID(db))).Methods("GET")
	r.HandleFunc("/{id}", logger(deleteImageByID(db))).Methods("DELETE")
	r.HandleFunc("", logger(queryImages(db))).Methods("GET")
	r.HandleFunc("", logger(createImage(db))).Methods("POST").Headers("Content-Type", "application/json")
}

func queryImages(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locator := repository.ImageLocator{}

		imageID, ok := r.URL.Query()["image_id"]
		if ok && len(imageID) >= 1 {
			locator.ImageID = imageID[0]
		}
		flavorID, ok := r.URL.Query()["flavor_id"]
		if ok && len(flavorID) >= 1 {
			locator.FlavorID = flavorID[0]
		}

		images, err := repository.GetImageFlavorRepository(db).RetrieveByFilterCriteria(locator)
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

func getImageByID(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := mux.Vars(r)["id"]
		exists, err := repository.GetImageFlavorRepository(db).RetrieveByUUID(uuid)
		if err == nil {
			if exists {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func deleteImageByID(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := mux.Vars(r)["id"]
		err := repository.GetImageFlavorRepository(db).DeleteByUUID(uuid)
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

func createImage(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var formBody model.Image
		if err := json.NewDecoder(r.Body).Decode(&formBody); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := repository.GetImageFlavorRepository(db).Create(&formBody); err != nil {
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
