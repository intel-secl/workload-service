package resource

import (
	"encoding/json"
	"fmt"
	"intel/isecl/lib/common/validation"
	flavorUtil "intel/isecl/lib/flavor/util"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// SetFlavorsEndpoints
func SetFlavorsEndpoints(r *mux.Router, db repository.WlsDatabase) {
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", errorHandler(getFlavorByID(db))).Methods("GET")
	r.HandleFunc("/{label}", errorHandler(getFlavorByLabel(db))).Methods("GET")
	r.HandleFunc("", (errorHandler(getFlavors(db)))).Methods("GET")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", errorHandler(deleteFlavorByID(db))).Methods("DELETE")
	r.HandleFunc("/{badid}", badId).Methods("DELETE")
	r.HandleFunc("", errorHandler(createFlavor(db))).Methods("POST").Headers("Content-Type", "application/json")
}

func getFlavorByID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		// validate uuid format
		if err := validation.ValidateUUIDv4(id); err != nil {
			log.Error("Invalid UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		fr := db.FlavorRepository()
		flavor, err := fr.RetrieveByUUID(id)
		uuidLog := log.WithField("uuid", id)
		if err != nil {
			log.WithError(err).Info("Failed to retrieve flavor by UUID")
			return err
		}
		if err := json.NewEncoder(w).Encode(flavor); err != nil {
			uuidLog.WithField("flavor", flavor).WithError(err).Error("Failed to encode JSON Flavor document")
			return err
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		uuidLog.Debug("Successfully fetched Flavor")
		return nil
	}
}

func getFlavorByLabel(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		label := mux.Vars(r)["label"]
		// validate label
		labelArr := []string{label}
		if validateInputErr := validation.ValidateStrings(labelArr); validateInputErr != nil {
			log.Error("Invalid label string format")
			return &endpointError{Message: validateInputErr.Error(), StatusCode: http.StatusBadRequest}
		}

		flavor, err := db.FlavorRepository().RetrieveByLabel(label)
		lblLog := log.WithField("label", label)
		if err != nil {
			lblLog.WithError(err).Info("Failed to retrieve Flavor by Label")
			return err
		}
		if err := json.NewEncoder(w).Encode(flavor); err != nil {
			lblLog.WithField("flavor", flavor).WithError(err).Error("Failed to encode JSON Flavor document")
			return err
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		lblLog.Debug("Successfully fetched Flavor")
		return nil
	}
}

func getFlavors(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		filterCriteria := repository.FlavorFilter{}
		flavorID, ok := r.URL.Query()["id"]
		if ok && len(flavorID) >= 1 {
			// validate UUID
			if err := validation.ValidateUUIDv4(flavorID[0]); err != nil {
				log.Error("Invalid UUID format")
				return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
			}
			filterCriteria.FlavorID = flavorID[0]
		}

		label, ok := r.URL.Query()["label"]
		if ok && len(label) >= 1 {
			// validate label string
			labelArr := []string{label[0]}
			if validateInputErr := validation.ValidateStrings(labelArr); validateInputErr != nil {
				log.Error("Invalid label string format")
				return &endpointError{Message: validateInputErr.Error(), StatusCode: http.StatusBadRequest}
			}
			filterCriteria.Label = label[0]
		}

		filter, ok := r.URL.Query()["filter"]
		if ok && len(filter) >= 1 {
			boolValue, err := strconv.ParseBool(filter[0])
			if err != nil {
				log.Error("Invalid filter boolean value, must be true or false")
				return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
			}
			filterCriteria.Filter = boolValue
		}

		flavors, err := db.FlavorRepository().RetrieveByFilterCriteria(filterCriteria)
		if err != nil {
			log.WithError(err).Info("Failed to retrieve flavors")
			return err
		}

		if err := json.NewEncoder(w).Encode(flavors); err != nil {
			log.WithError(err).Error("Unexpectedly failed to encode flavors to JSON")
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
		// validate uuid format
		if err := validation.ValidateUUIDv4(id); err != nil {
			log.Error("Invalid UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}

		fr := db.FlavorRepository()
		uuidLog := log.WithField("uuid", id)
		if err := fr.DeleteByUUID(id); err != nil {
			uuidLog.WithError(err).Info("Failed to delete Flavor by UUID")
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		uuidLog.Debug("Successfully deleted Flavor")
		return nil
	}
}

func createFlavor(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var f flavorUtil.SignedImageFlavor
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&f); err != nil {
			log.WithError(err).Info("Failed to encode request body as Flavor")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}

		if f.ImageFlavor.Meta.Description.FlavorPart == "" ||
			(f.ImageFlavor.Meta.Description.FlavorPart != "CONTAINER_IMAGE" && f.ImageFlavor.Meta.Description.FlavorPart != "IMAGE") {
			msg := fmt.Sprintf("Invalid FlavorPart value: %s", f.ImageFlavor.Meta.Description.FlavorPart)
			log.Error(msg)
			return &endpointError{Message: msg, StatusCode: http.StatusBadRequest}
		}

		if f.Signature == "" {
			msg := fmt.Sprintf("Flavor signature not provided in input")
			log.Error(msg)
			return &endpointError{Message: msg, StatusCode: http.StatusBadRequest}
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
		fLog := log.WithField("flavor", f)
		switch err := fr.Create(&f); err {
		case repository.ErrFlavorLabelAlreadyExists:
			msg := fmt.Sprintf("Flavor with Label %s already exists", f.ImageFlavor.Meta.Description.Label)
			fLog.Info(msg)
			return &endpointError{
				Message:    msg,
				StatusCode: http.StatusConflict,
			}
		case repository.ErrFlavorUUIDAlreadyExists:
			msg := fmt.Sprintf("Flavor with UUID %s already exists", f.ImageFlavor.Meta.ID)
			fLog.Info(msg)
			return &endpointError{
				Message:    msg,
				StatusCode: http.StatusConflict,
			}
		case nil:
			w.WriteHeader(http.StatusCreated)
			var flavor model.Flavor
			flavor.Image = f.ImageFlavor
			if err := json.NewEncoder(w).Encode(flavor); err != nil {
				fLog.Error("Unexpectedly failed to encode Flavor to JSON")
				return err
			}
			fLog.Debug("Successfully created Flavor")
			return nil
		default:
			fLog.WithError(err).Error("Unexpected error when writing Flavor to Database")
			return &endpointError{
				Message:    "Unexpected error when writing Flavor to Database, check input format",
				StatusCode: http.StatusBadRequest,
			}
		}
	}
}
