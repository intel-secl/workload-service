package resource

import (
	"encoding/json"
	"fmt"
	"intel/isecl/workload-service/repository"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// SetImagesEndpoints sets endpoints for /image
func SetImagesEndpoints(r *mux.Router, db *gorm.DB) {
	r.HandleFunc("/{id}", getImageByID(db)).Methods("GET")
	r.HandleFunc("/{id}", deleteImageByID(db)).Methods("DELETE")
	//r.HandleFunc("", nil).Methods("GET")
	r.HandleFunc("", createImage(db)).Methods("POST").Headers("Content-Type", "application/json")
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

// CreateImage defines the request body for a POST /image request, creating an association between ImageID and
type CreateImage struct {
	ImageID  string `json:"image_id"`
	FlavorID string `json:"flavor_id"`
}

func createImage(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var formBody CreateImage
		if err := json.NewDecoder(r.Body).Decode(&formBody); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := repository.GetImageFlavorRepository(db).Create(formBody.ImageID, formBody.FlavorID); err != nil {
			switch err {
			case repository.ErrImageAssociationAlreadyExists:
				http.Error(w, fmt.Sprintf("image with UUID %s is already associated with a flavor", formBody.ImageID), http.StatusConflict)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		w.WriteHeader(http.StatusCreated)
	}
}
