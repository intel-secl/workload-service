/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package resource

import (
	"encoding/json"
	"fmt"
	"intel/isecl/lib/common/log/message"
	"intel/isecl/lib/common/validation"
	flvr "intel/isecl/lib/flavor"
	consts "intel/isecl/workload-service/constants"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// SetFlavorsEndpoints
func SetFlavorsEndpoints(r *mux.Router, db repository.WlsDatabase) {
	log.Trace("resource/flavors:SetFlavorsEndpoints() Entering")
	defer log.Trace("resource/flavors:SetFlavorsEndpoints() Leaving")
	r.HandleFunc("/{id:(?i:[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$)}",
		errorHandler(requiresPermission(getFlavorByID(db), []string{consts.AdministratorGroupName}))).Methods("GET")
	r.HandleFunc("/{label}", errorHandler(requiresPermission(getFlavorByLabel(db), []string{consts.AdministratorGroupName}))).Methods("GET")
	r.HandleFunc("", (errorHandler(requiresPermission(getFlavors(db), []string{consts.AdministratorGroupName})))).Methods("GET")
	r.HandleFunc("/{id:(?i:[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$)}",
		errorHandler(requiresPermission(deleteFlavorByID(db), []string{consts.AdministratorGroupName}))).Methods("DELETE")
	r.HandleFunc("", errorHandler(requiresPermission(createFlavor(db), []string{consts.AdministratorGroupName}))).Methods("POST").Headers("Content-Type", "application/json")
	r.HandleFunc("/{badid}", badId).Methods("DELETE")
}

func getFlavorByID(db repository.WlsDatabase) endpointHandler {
	log.Trace("resource/flavors:getFlavorByID() Entering")
	defer log.Trace("resource/flavors:getFlavorByID() Leaving")
	return func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		// validate uuid format
		if err := validation.ValidateUUIDv4(id); err != nil {
			seclog.Errorf("resource/flavors:getFlavorByID() %s : Invalid UUID format - %s", message.InvalidInputBadParam, id)
			log.Tracef("%+v", err)
			return &endpointError{Message: "Failed to retrieve flavor - Invalid UUID format", StatusCode: http.StatusBadRequest}
		}
		fr := db.FlavorRepository()
		flavor, err := fr.RetrieveByUUID(id)
		uuidLog := log.WithField("uuid", id)
		if err != nil {
			uuidLog.WithError(err).Errorf("resource/flavors:getFlavorByID() %s : Failed to retrieve flavor by UUID", message.AppRuntimeErr)
			log.Error(message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{Message: "Failed to retrieve flavor by UUID - Record not found", StatusCode: http.StatusNotFound}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(flavor); err != nil {
			uuidLog.WithField("flavor", flavor).WithError(err).Errorf("resource/flavors:getFlavorByID() %s : JSON Flavor document encode failure", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{Message: "Failed to retrieve flavor by UUID - JSON marshal failure", StatusCode: http.StatusInternalServerError}
		}
		uuidLog.Info("resource/flavors:getFlavorByID() Successfully fetched Flavor")
		return nil
	}
}

func getFlavorByLabel(db repository.WlsDatabase) endpointHandler {
	log.Trace("resource/flavors:getFlavorByLabel() Entering")
	defer log.Trace("resource/flavors:getFlavorByLabel() Leaving")
	return func(w http.ResponseWriter, r *http.Request) error {
		label := mux.Vars(r)["label"]
		// validate label
		labelArr := []string{label}
		if validateInputErr := validation.ValidateStrings(labelArr); validateInputErr != nil {
			log.Errorf("resource/flavors:getFlavorByLabel() %s : Invalid label string format", message.InvalidInputProtocolViolation)
			return &endpointError{Message: "Failed to retrieve flavor by label - Invalid label string format", StatusCode: http.StatusBadRequest}
		}

		flavor, err := db.FlavorRepository().RetrieveByLabel(label)
		lblLog := log.WithField("label", label)
		if err != nil {
			lblLog.WithError(err).Errorf("resource/flavors:getFlavorByLabel() %s : Failed to retrieve Flavor by Label", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{Message: "Failed to retrieve flavor by label - backend error", StatusCode: http.StatusInternalServerError}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(flavor); err != nil {
			lblLog.WithField("flavor", flavor).WithError(err).Errorf("resource/flavors:getFlavorByLabel() %s : JSON Encode error", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{Message: "Failed to retrieve flavor by label - JSON Encode error", StatusCode: http.StatusInternalServerError}
		}
		lblLog.Info("resource/flavors:getFlavorByLabel() Successfully fetched Flavor")
		return nil
	}
}

func getFlavors(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/flavors:getFlavors() Entering")
		defer log.Trace("resource/flavors:getFlavors() Leaving")
		var fLog = log
		filterCriteria := repository.FlavorFilter{}
		flavorID, ok := r.URL.Query()["id"]

		if ok && len(flavorID[0]) >= 1 {
			// validate UUID
			if err := validation.ValidateUUIDv4(flavorID[0]); err != nil {
				log.Errorf("resource/flavors:getFlavors() %s : Invalid flavor UUID format", message.InvalidInputProtocolViolation)
				log.Tracef("%+v", err)
				return &endpointError{Message: "Unable to retrieve flavor - Invalid flavor UUID format", StatusCode: http.StatusBadRequest}
			}
			filterCriteria.FlavorID = flavorID[0]
			fLog = log.WithField("flavorid", flavorID[0])
		}

		label, ok := r.URL.Query()["label"]
		if ok && len(label[0]) >= 1 {
			// validate label string
			labelArr := []string{label[0]}
			if validateInputErr := validation.ValidateStrings(labelArr); validateInputErr != nil {
				log.Errorf("resource/flavors:getFlavors() %s : Invalid label string format", message.InvalidInputProtocolViolation)
				return &endpointError{Message: "Unable to retrieve flavor - Invalid label string", StatusCode: http.StatusBadRequest}
			}
			filterCriteria.Label = label[0]
			fLog = fLog.WithField("label", label[0])
		}

		filter, ok := r.URL.Query()["filter"]
		if ok && len(filter[0]) >= 1 {
			boolValue, err := strconv.ParseBool(filter[0])
			if err != nil {
				fLog.WithError(err).Errorf("resource/flavors:getFlavors() %s : Invalid filter boolean value, must be true or false", message.InvalidInputProtocolViolation)
				log.Tracef("%+v", err)
				return &endpointError{Message: "Unable to retrieve flavor - Invalid filter boolean value, must be true or false", StatusCode: http.StatusBadRequest}
			}
			filterCriteria.Filter = boolValue
		}

		if filterCriteria.Label == "" && filterCriteria.FlavorID == "" && filterCriteria.Filter {
			log.Errorf("resource/flavors:getFlavors() %s : Invalid filter criteria. Allowed filter critierias are id, label and filter = false\n", message.InvalidInputProtocolViolation)
			return &endpointError{Message: "Unable to retrieve flavor - Invalid filter criteria - Allowed filter critierias are id, label and filter = false", StatusCode: http.StatusBadRequest}
		}

		flavors, err := db.FlavorRepository().RetrieveByFilterCriteria(filterCriteria)
		if err != nil {
			fLog.WithError(err).Errorf("resource/flavors:getFlavors() %s : Failed to retrieve flavors", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{Message: "Unable to retrieve flavor - backend error", StatusCode: http.StatusInternalServerError}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(flavors); err != nil {
			fLog.WithError(err).Errorf("resource/flavors:getFlavors() %s : Unexpectedly failed to encode flavors to JSON", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{Message: "Unable to retrieve flavor - Failed to encode flavors to JSON", StatusCode: http.StatusInternalServerError}
		}
		fLog.Info("resource/flavors:getFlavors() Successfully fetched Flavor")
		return nil
	}
}

func deleteFlavorByID(db repository.WlsDatabase) endpointHandler {
	log.Trace("resource/flavors:deleteFlavorByID() Entering")
	defer log.Trace("resource/flavors:deleteFlavorByID() Leaving")
	return func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		// validate uuid format
		if err := validation.ValidateUUIDv4(id); err != nil {
			log.Errorf("resource/flavors:deleteFlavorByID() %s : Invalid UUID format", message.InvalidInputProtocolViolation)
			log.Tracef("%+v", err)
			return &endpointError{Message: "Failed to delete flavor - Invalid UUID", StatusCode: http.StatusBadRequest}
		}

		fr := db.FlavorRepository()
		uuidLog := log.WithField("uuid", id)
		if err := fr.DeleteByUUID(id); err != nil {

			if err.Error() == "record not found" {
				return &endpointError{Message: "Non-existent flavor", StatusCode: http.StatusNotFound}
			}
			uuidLog.WithError(err).Errorf("resource/flavors:deleteFlavorByID() %s : Failed to delete Flavor by UUID", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{Message: "Failed to delete flavor", StatusCode: http.StatusInternalServerError}
		}
		w.WriteHeader(http.StatusNoContent)
		uuidLog.Info("resource/flavors:deleteFlavorByID() Successfully deleted Flavor")
		return nil
	}
}

func createFlavor(db repository.WlsDatabase) endpointHandler {
	log.Trace("resource/flavors:createFlavor() Entering")
	defer log.Trace("resource/flavors:createFlavor() Leaving")

	return func(w http.ResponseWriter, r *http.Request) error {
		var f flvr.SignedImageFlavor
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&f); err != nil {
			log.WithError(err).Errorf("resource/flavors:createFlavor() %s :  Failed to encode request body as Flavor", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{Message: "Failed to delete flavor", StatusCode: http.StatusBadRequest}
		}

		if f.ImageFlavor.Meta.Description.FlavorPart == "" ||
			(f.ImageFlavor.Meta.Description.FlavorPart != "CONTAINER_IMAGE" && f.ImageFlavor.Meta.Description.FlavorPart != "IMAGE") {
			msg := fmt.Sprintf("Invalid FlavorPart value: %s", f.ImageFlavor.Meta.Description.FlavorPart)
			log.Errorf("resource/flavors:createFlavor() %s : Failed to create flavor: "+msg, message.AppRuntimeErr)
			return &endpointError{Message: msg, StatusCode: http.StatusBadRequest}
		}

		if f.Signature == "" {
			msg := fmt.Sprintf("Flavor signature not provided in input")
			log.Errorf("resource/flavors:createFlavor() %s : Failed to create flavor: "+msg, message.InvalidInputBadParam)
			return &endpointError{Message: msg, StatusCode: http.StatusBadRequest}
		}

		// it's almost silly that we unmarshal, then remarshal it to store it back into the database, // but at least it provides some validation of the input
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
			fLog.Errorf("resource/flavors:createFlavor() %s : "+msg, message.InvalidInputProtocolViolation)
			return &endpointError{
				Message:    msg,
				StatusCode: http.StatusConflict,
			}
		case repository.ErrFlavorUUIDAlreadyExists:
			msg := fmt.Sprintf("Flavor with UUID %s already exists", f.ImageFlavor.Meta.ID)
			fLog.Errorf("resource/flavors:createFlavor() %s : "+msg, message.InvalidInputProtocolViolation)
			return &endpointError{
				Message:    msg,
				StatusCode: http.StatusConflict,
			}
		case nil:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			var flavor model.Flavor
			flavor.Image = f.ImageFlavor
			if err := json.NewEncoder(w).Encode(flavor); err != nil {
				fLog.WithError(err).Errorf("resource/flavors:createFlavor() %s : Unexpectedly failed to encode Flavor to JSON", message.AppRuntimeErr)
				log.Tracef("%+v", err)
				return &endpointError{Message: "Failed to create flavor - JSON encode failed", StatusCode: http.StatusInternalServerError}
			}
			fLog.Debug("resource/flavors:createFlavor() Successfully created Flavor")
			return nil
		default:
			fLog.WithError(err).Errorf("resource/flavors:createFlavor() %s : Unexpected error when writing Flavor to Database", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Unexpected error when writing Flavor to Database, check input format",
				StatusCode: http.StatusBadRequest,
			}
		}
	}
}
