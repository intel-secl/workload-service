/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package resource

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"intel/isecl/lib/common/v3/crypt"
	"intel/isecl/lib/common/v3/log/message"
	"intel/isecl/lib/common/v3/validation"
	"intel/isecl/workload-service/v3/constants"
	consts "intel/isecl/workload-service/v3/constants"
	"intel/isecl/workload-service/v3/keycache"
	"intel/isecl/workload-service/v3/model"
	"intel/isecl/workload-service/v3/repository"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Saml is used to represent saml report struct
type Saml struct {
	XMLName   xml.Name    `xml:"Assertion"`
	Subject   Subject     `xml:"Subject>SubjectConfirmation>SubjectConfirmationData"`
	Attribute []Attribute `xml:"AttributeStatement>Attribute"`
	Signature string      `xml:"Signature>KeyInfo>X509Data>X509Certificate"`
}

type Subject struct {
	XMLName      xml.Name  `xml:"SubjectConfirmationData"`
	NotBefore    time.Time `xml:"NotBefore,attr"`
	NotOnOrAfter time.Time `xml:"NotOnOrAfter,attr"`
}

type Attribute struct {
	XMLName        xml.Name `xml:"Attribute"`
	Name           string   `xml:"Name,attr"`
	AttributeValue string   `xml:"AttributeValue"`
}

// SetImagesEndpoints sets endpoints for /image
func SetImagesEndpoints(r *mux.Router, db repository.WlsDatabase) {
	log.Trace("resource/images:SetImagesEndpoints() Entering")
	defer log.Trace("resource/images:SetImagesEndpoints() Leaving")
	//There is a ambiguity between api endpoints /<id>/flavors and /<id>/flavors?flavor_part=<flavor_part>
	//Moved /<id>/flavors?flavor_part=<flavor_part> to top so this will be able to check for the filter flavor_part
	r.HandleFunc("/{id}/flavors",
		errorHandler(requiresPermission(retrieveFlavorForImageID(db), []string{constants.ImageFlavorsRetrieve}))).Methods("GET").Queries("flavor_part", "{flavor_part}")
	r.HandleFunc("/{id}/flavors",
		errorHandler(requiresPermission(getAllAssociatedFlavors(db), []string{constants.ImageFlavorsSearch}))).Methods("GET")
	r.HandleFunc("/{id}/flavors/{flavorID}",
		errorHandler(requiresPermission(getAssociatedFlavor(db), []string{constants.ImageFlavorsRetrieve}))).Methods("GET")
	r.HandleFunc("/{id}/flavors/{flavorID}",
		errorHandler(requiresPermission(putAssociatedFlavor(db), []string{constants.ImageFlavorsStore}))).Methods("PUT")
	r.HandleFunc("/{id}/flavors/{flavorID}",
		errorHandler(requiresPermission(deleteAssociatedFlavor(db), []string{constants.ImageFlavorsDelete}))).Methods("DELETE")
	r.HandleFunc("/{id}",
		errorHandler(requiresPermission(getImageByID(db), []string{constants.ImagesRetrieve}))).Methods("GET")
	r.HandleFunc("/{id}",
		errorHandler(requiresPermission(deleteImageByID(db), []string{constants.ImagesDelete}))).Methods("DELETE")
	r.HandleFunc("/{id}/flavor-key",
		errorHandler(requiresPermission(retrieveFlavorAndKeyForImageID(db), []string{constants.ImageFlavorsRetrieve}))).Methods("GET").Queries("hardware_uuid", "{hardware_uuid}")
	r.HandleFunc("/{id}/flavor-key",
		missingQueryParameters("hardware_uuid")).Methods("GET")
	r.HandleFunc("",
		errorHandler(requiresPermission(queryImages(db), []string{constants.ImagesSearch}))).Methods("GET")
	r.HandleFunc("",
		errorHandler(requiresPermission(createImage(db), []string{constants.ImagesCreate}))).Methods("POST").Headers("Content-Type", "application/json")
	r.HandleFunc("/{badid}", badId)
}

// Logs error if a UUID is not in UUIDv4 format
func badId(w http.ResponseWriter, r *http.Request) {
	log.Trace("resource/images:badId() Entering")
	defer log.Trace("resource/images:badId() Leaving")
	badid := mux.Vars(r)["badid"]
	log.Errorf("resource/images:badId() %s : Request made with non compliant UUIDv4: %v", message.InvalidInputProtocolViolation, badid)

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintf("%s is not uuidv4 compliant", badid)))
}

// Logs error if a query is missing one or more parameters
func missingQueryParameters(params ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Trace("resource/images:missingQueryParameters() Entering")
		defer log.Trace("resource/images:missingQueryParameters() Leaving")
		errStr := fmt.Sprintf("Missing query parameters: %v", params)
		log.Errorf("resource/images:missingQueryParameters() %s : %s", message.InvalidInputBadParam, errStr)
		http.Error(w, errStr, http.StatusBadRequest)
	}
}

// Retrieves the flavor and key corresponding to a image UUID
func retrieveFlavorAndKeyForImageID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/images:retrieveFlavorAndKeyForImageID() Entering")
		defer log.Trace("resource/images:retrieveFlavorAndKeyForImageID() Leaving")

		id := mux.Vars(r)["id"]
		// validate UUID format
		if err := validation.ValidateUUIDv4(id); err != nil {
			log.Errorf("resource/images:retrieveFlavorAndKeyForImageID() %s : Invalid UUID format - %s", message.InvalidInputProtocolViolation, id)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve Flavor/Key - Invalid UUID",
				StatusCode: http.StatusBadRequest,
			}
		}
		hwid := mux.Vars(r)["hardware_uuid"]
		// validate hardware UUID
		if err := validation.ValidateHardwareUUID(hwid); err != nil {
			log.Errorf("resource/images:retrieveFlavorAndKeyForImageID() %s : Invalid hardware UUID format - %s", message.InvalidInputProtocolViolation, hwid)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve Flavor/Key - Invalid hardware uuid",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("imageUUID", id).WithField("hardwareUUID", hwid)

		cLog.Debug("resource/images:retrieveFlavorAndKeyForImageID() Retrieving Flavor and Key for Image")
		flavor, err := db.ImageRepository().RetrieveAssociatedImageFlavor(id)
		if err != nil {
			cLog.WithError(err).Errorf("resource/images:retrieveFlavorAndKeyForImageID() %s : Failed to retrieve Flavor and Key for Image", message.AppRuntimeErr)
			return &endpointError{
				Message:    "Failed to retrieve Flavor and Key for Image - Backend Error",
				StatusCode: http.StatusNotFound,
			}
		}

		keyUrl := flavor.ImageFlavor.Encryption.KeyURL
		// Check if flavor keyUrl is not empty
		if flavor.ImageFlavor.EncryptionRequired && len(flavor.ImageFlavor.Encryption.KeyURL) > 0 {
			key, err := transfer_key(true, hwid, keyUrl, id)
			if err != nil {
                                cLog.WithError(err).Error("resource/images:retrieveFlavorAndKeyForImageID() Error while retrieving key")
				return err
			}

			// got key data
			flavorKey := model.FlavorKey{
				Flavor:    flavor.ImageFlavor,
				Signature: flavor.Signature,
				Key:       key,
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(flavorKey); err != nil {
				// marshalling error 500
				cLog.WithError(err).Errorf("resource/images:retrieveFlavorAndKeyForImageID() %s : Unexpectedly failed to encode FlavorKey to JSON", message.AppRuntimeErr)
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "Failed to encode FlavorKey to JSON - Failure marshalling JSON response",
					StatusCode: http.StatusInternalServerError,
				}
			}
			cLog.Info("resource/images:retrieveFlavorAndKeyForImageID() Successfully retrieved FlavorKey")
			return nil
		}
		// just return the flavor
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(model.FlavorKey{Flavor: flavor.ImageFlavor, Signature: flavor.Signature}); err != nil {
			// marshalling error 500
			cLog.WithError(err).Errorf("resource/images:retrieveFlavorAndKeyForImageID() %s : Unexpectedly failed to encode FlavorKey to JSON", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve FlavorKey - Failure marshalling JSON response",
				StatusCode: http.StatusInternalServerError,
			}
		}
		cLog.Info("resource/images:retrieveFlavorAndKeyForImageID() Successfully retrieved Flavor and Key")
		return nil
	}
}

func retrieveFlavorForImageID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/images:retrieveFlavorForImageID() Entering")
		defer log.Trace("resource/images:retrieveFlavorForImageID() Leaving")
		id := mux.Vars(r)["id"]
		// validate UUID
		if err := validation.ValidateUUIDv4(id); err != nil {
			log.WithError(err).Error("resource/images:retrieveFlavorForImageID() Invalid UUID format")
			return &endpointError{
				Message:    "Failed to retrieve flavor - Invalid image UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		fp := mux.Vars(r)["flavor_part"]
		// validate flavor part
		fpArr := []string{fp}
		if validateInputErr := validation.ValidateStrings(fpArr); validateInputErr != nil {
			log.WithError(validateInputErr).Errorf("resource/images:retrieveFlavorForImageID() %s : Invalid flavor part string format", message.InvalidInputProtocolViolation)
			return &endpointError{
				Message:    "Failed to retrieve flavor - Invalid flavor part string format",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("imageUUID", id).WithField("flavorPart", fp)

		if fp == "" {
			cLog.Errorf("resource/images:retrieveFlavorForImageID() %s : Missing required parameter flavor_part", message.InvalidInputBadParam)
			return &endpointError{
				Message:    "Failed to retrieve flavor - Query parameter 'flavor_part' cannot be nil",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog.Info("resource/images:retrieveFlavorForImageID() Retrieving Flavor for Image")
		// The error returned for below db query should be set as "StatusNotFound" for http response status code,
		// since image UUID and flavorUUID validation have been done earlier in the code
		// there would be rare case when there is flavor in db and query fails to fetch flavor

		flavor, err := db.ImageRepository().RetrieveAssociatedFlavorByFlavorPart(id, fp)
		if err != nil {
			cLog.WithError(err).Errorf("resource/images:retrieveFlavorForImageID() %s : Failed to retrieve Flavor for Image", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve flavor - No flavor found for given image ID",
				StatusCode: http.StatusNotFound,
			}
		}

		// just return the flavor
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(*flavor); err != nil {
			// marshalling error 500
			cLog.WithError(err).Errorf("resource/images:retrieveFlavorForImageID() %s : Unexpectedly failed to encode FlavorKey to JSON", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve flavor - Failure marshal FlavorKey JSON response",
				StatusCode: http.StatusInternalServerError,
			}
		}
		return nil
	}
}

func getAllAssociatedFlavors(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/images:getAllAssociatedFlavors() Entering")
		defer log.Trace("resource/images:getAllAssociatedFlavors() Leaving")

		uuid := mux.Vars(r)["id"]
		// validate UUID format
		if err := validation.ValidateUUIDv4(uuid); err != nil {
			log.WithError(err).Errorf("resource/images:getAllAssociatedFlavors() %s : Invalid UUID format", message.InvalidInputProtocolViolation)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve associated flavors - Invalid UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("uuid", uuid)
		flavors, err := db.ImageRepository().RetrieveAssociatedFlavors(uuid)
		if err != nil {
			if strings.Contains(err.Error(), "record not found") {
				cLog.WithError(err).Errorf("resource/images:getAllAssociatedFlavors() %s : Failed to retrieve associated flavors for image", message.AppRuntimeErr)
				log.Tracef("%+v", err)
				log.Debug(err.Error())
				return &endpointError{
					Message:    "Failed to retrieve associated flavors - No Flavor found for Image",
					StatusCode: http.StatusNotFound,
				}
			} else {
				cLog.WithError(err).Errorf("resource/images:getAllAssociatedFlavors() %s : Failed to retrieve associated flavors for image", message.AppRuntimeErr)
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "Failed to retrieve associated flavors - backend error",
					StatusCode: http.StatusInternalServerError,
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(flavors); err != nil {
			cLog.WithError(err).Errorf("resource/images:getAllAssociatedFlavors() %s : Unexpectedly failed to encode list of flavors to JSON", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve associated flavors - JSON marshal response failure",
				StatusCode: http.StatusInternalServerError,
			}
		}
		cLog.WithField("flavors", flavors).Info("resource/images:getAllAssociatedFlavors() Successfully retrieved associated flavors for image")
		return nil
	}
}

func getAssociatedFlavor(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/images:getAssociatedFlavor() Entering")
		defer log.Trace("resource/images:getAssociatedFlavor() Leaving")

		imageUUID := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(imageUUID); err != nil {
			log.WithError(err).Errorf("resource/images:getAssociatedFlavor() %s : Invalid image UUID format", message.InvalidInputProtocolViolation)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve flavor - invalid image UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		flavorUUID := mux.Vars(r)["flavorID"]
		// validate flavor UUID
		if err := validation.ValidateUUIDv4(flavorUUID); err != nil {
			log.WithError(err).Errorf("resource/images:getAssociatedFlavor() %s : Invalid flavor UUID format", message.InvalidInputProtocolViolation)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve associated flavor for image - Invalid image UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("imageUUID", imageUUID).WithField("flavorUUID", flavorUUID)

		flavor, err := db.ImageRepository().RetrieveAssociatedFlavor(imageUUID, flavorUUID)
		if err != nil {
			if strings.Contains(err.Error(), "record not found") {
				cLog.WithError(err).Errorf("resource/images:getAssociatedFlavor() %s : Failed to retrieve associated flavors for image", message.AppRuntimeErr)
				log.Debug(err.Error())
				return &endpointError{
					Message:    "Failed to retrieve associated flavors - No flavor associated with given image UUID",
					StatusCode: http.StatusNotFound,
				}
			} else {
				cLog.WithError(err).Errorf("resource/images:getAssociatedFlavor() %s : Failed to retrieve associated flavors for image", message.AppRuntimeErr)
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "Failed to retrieve associated flavors - backend error",
					StatusCode: http.StatusInternalServerError,
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		cLog = cLog.WithField("flavor", flavor)
		if err := json.NewEncoder(w).Encode(flavor); err != nil {
			cLog.WithError(err).Errorf("resource/images:getAssociatedFlavor() %s : Unexpectedly failed to encode Flavor to JSON", message.AppRuntimeErr)
			return &endpointError{
				Message:    "Failed to retrieve associated flavor for image - JSON response encode marshal failure",
				StatusCode: http.StatusInternalServerError,
			}
		}
		cLog.Info("resource/images:getAssociatedFlavor() Successfully retrieved associated Flavor")
		return nil
	}
}

func putAssociatedFlavor(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/images:putAssociatedFlavor() Entering")
		defer log.Trace("resource/images:putAssociatedFlavor() Leaving")

		imageUUID := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(imageUUID); err != nil {
			log.WithError(err).Error("resource/images:putAssociatedFlavor() Invalid image UUID format")
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to create image/flavor association - invalid image UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		flavorUUID := mux.Vars(r)["flavorID"]
		// validate flavor UUID
		if err := validation.ValidateUUIDv4(flavorUUID); err != nil {
			log.WithError(err).Errorf("resource/images:putAssociatedFlavor() %s : Invalid flavor UUID format", message.InvalidInputProtocolViolation)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to create image/flavor association - invalid flavor UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("imageUUID", imageUUID).WithField("flavorUUID", flavorUUID)

		if err := db.ImageRepository().AddAssociatedFlavor(imageUUID, flavorUUID); err != nil {
			cLog.WithError(err).Errorf("resource/images:putAssociatedFlavor() %s : Failed to add new Flavor association", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			if strings.Contains(err.Error(), "record not found") {
				return &endpointError{
					Message:    "Failed to create image/flavor association - Record not found",
					StatusCode: http.StatusNotFound,
				}
			}
			return &endpointError{
				Message:    "Failed to create image/flavor association - Backend error",
				StatusCode: http.StatusInternalServerError,
			}
		}
		w.WriteHeader(http.StatusCreated)
		cLog.Info("resource/images:putAssociatedFlavor() Successfully added new Flavor association")
		return nil
	}
}

func deleteAssociatedFlavor(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/images:deleteAssociatedFlavor() Entering")
		defer log.Trace("resource/images:deleteAssociatedFlavor() Leaving")

		imageUUID := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(imageUUID); err != nil {
			log.WithError(err).Errorf("resource/images:deleteAssociatedFlavor() %s : Invalid image UUID format", message.InvalidInputProtocolViolation)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to delete image/flavor association - invalid image UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		flavorUUID := mux.Vars(r)["flavorID"]
		// validate flavor UUID
		if err := validation.ValidateUUIDv4(flavorUUID); err != nil {
			log.WithError(err).Errorf("resource/images:deleteAssociatedFlavor() %s : Invalid flavor UUID format", message.InvalidInputProtocolViolation)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to delete image/flavor association - invalid flavor UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("imageUUID", imageUUID).WithField("flavorUUID", flavorUUID)
		err := db.ImageRepository().DeleteAssociatedFlavor(imageUUID, flavorUUID)
		if err != nil {
			cLog.WithError(err).Errorf("resource/images:deleteAssociatedFlavor() %s : Failed to remove Flavor association for Image", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to delete image/flavor association - Backend error",
				StatusCode: http.StatusInternalServerError,
			}
		}
		w.WriteHeader(http.StatusNoContent)
		cLog.Debug("resource/images:deleteAssociatedFlavor() Successfully removed Flavor association for Image")
		return nil
	}
}

// wls/images --> (w/o params) return 400 and error message
// wls/images?filter=false --> all images in db, status ok
// wls/images?flavor_id --> filter on flavor id, status ok
// wls/images?image_id --> filter on image id, status ok
// all other parameter options --> 400 with error message
func queryImages(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/images:queryImages() Entering")
		defer log.Trace("resource/images:queryImages() Leaving")
		var cLog = log

		locator := repository.ImageFilter{}
		locator.Filter = true // default to 'filter' to true

		if len(r.URL.Query()) == 0 {
			http.Error(w, "At least one query parameter is required", http.StatusBadRequest)
			return nil
		}

		filter, ok := r.URL.Query()["filter"]
		if ok && len(filter[0]) >= 1 {
			boolValue, err := strconv.ParseBool(filter[0])
			if err != nil {
				log.WithError(err).Errorf("resource/images:queryImages() %s : Invalid filter boolean value, must be true or false", message.InvalidInputProtocolViolation)
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "Failed to retrieve image - Invalid filter boolean value, must be true or false",
					StatusCode: http.StatusBadRequest,
				}
			}
			locator.Filter = boolValue
			cLog = cLog.WithField("Filter", boolValue)
		}

		flavorID, ok := r.URL.Query()["flavor_id"]
		if ok && len(flavorID[0]) >= 1 {
			if err := validation.ValidateUUIDv4(flavorID[0]); err != nil {
				cLog.WithError(err).Errorf("resource/images:queryImages() %s : Invalid flavor UUID format", message.InvalidInputProtocolViolation)
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "Failed to retrieve image - Invalid flavor UUID format",
					StatusCode: http.StatusBadRequest,
				}
			}
			locator.FlavorID = flavorID[0]
			cLog = cLog.WithField("FlavorID", flavorID[0])
		}

		imageID, ok := r.URL.Query()["image_id"]
		if ok && len(imageID[0]) >= 1 {
			if err := validation.ValidateUUIDv4(imageID[0]); err != nil {
				cLog.WithError(err).Errorf("resource/images:queryImages() %s : Invalid image UUID format", message.InvalidInputProtocolViolation)
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "Failed to retrieve image - Invalid image UUID format",
					StatusCode: http.StatusBadRequest,
				}
			}
			locator.ImageID = imageID[0]
			cLog = cLog.WithField("id", imageID[0])
		}

		if locator.FlavorID == "" && locator.ImageID == "" && locator.Filter {
			cLog.Errorf("resource/images:queryImages() %s : Invalid filter criteria. Allowed filter critierias are image_id, flavor_id and filter = false\n", message.InvalidInputBadParam)
			return &endpointError{
				Message:    "Failed to retrieve image - Invalid filter criteria. Allowed filter critierias are image_id, flavor_id and filter = false",
				StatusCode: http.StatusBadRequest,
			}
		}

		images, err := db.ImageRepository().RetrieveByFilterCriteria(locator)
		if err != nil {
			cLog.WithError(err).Errorf("resource/images:queryImages() %s : Failed to retrieve Images by filter criteria", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve image - Failed to retrieve Images by filter criteria",
				StatusCode: http.StatusInternalServerError,
			}
		}
		if images == nil {
			// coerce to return empty list instead of null
			images = []model.Image{}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(images); err != nil {
			cLog.WithError(err).Errorf("resource/images:queryImages() %s : Unexpectedly failed to encode array of Images to JSON", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve image - JSON response encode marshal failure",
				StatusCode: http.StatusInternalServerError,
			}
		}
		cLog.Info("resource/images:queryImages() Successfully queried Images by filter criteria")
		return nil
	}
}

func getImageByID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/images:getImageByID() Entering")
		defer log.Trace("resource/images:getImageByID() Leaving")

		uuid := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(uuid); err != nil {
			log.WithError(err).Errorf("resource/images:getImageByID() %s : Invalid image UUID format", message.InvalidInputProtocolViolation)
			log.Tracef("%+v", err)
			return &endpointError{Message: "Failed to retrieve image - Invalid image UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("uuid", uuid)
		image, err := db.ImageRepository().RetrieveByUUID(uuid)
		if err != nil {
			cLog.WithError(err).Errorf("resource/images:getImageByID() %s : Failed to retrieve Image by UUID", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "No image found for given UUID",
				StatusCode: http.StatusNotFound,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		cLog = cLog.WithField("image", image)
		if err := json.NewEncoder(w).Encode(image); err != nil {
			cLog.WithError(err).Errorf("resource/images:getImageByID() %s : Unexpectedly failed to encode Image to JSON", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve image - JSON response encode marshal failure",
				StatusCode: http.StatusInternalServerError,
			}
		}
		cLog.Info("resource/images:getImageByID() Successfully retrieved Image by UUID")
		return nil
	}
}

func deleteImageByID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/images:deleteImageByID() Entering")
		defer log.Trace("resource/images:deleteImageByID() Leaving")

		uuid := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(uuid); err != nil {
			log.WithError(err).Errorf("resource/images:deleteImageByID() %s : Invalid image UUID format", message.InvalidInputProtocolViolation)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to delete image - invalid image UUID",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("uuid", uuid)
		if err := db.ImageRepository().DeleteByUUID(uuid); err != nil {
			cLog.WithError(err).Errorf("resource/images:deleteImageByID() %s : Failed to delete Image by UUID", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to delete image - backend error",
				StatusCode: http.StatusInternalServerError,
			}
		}
		w.WriteHeader(http.StatusNoContent)
		cLog.Info("resource/images:deleteImageByID() Successfully deleted Image by UUID")
		return nil
	}
}

func createImage(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/images:createImage() Entering")
		defer log.Trace("resource/images:createImage() Leaving")

		var formBody model.Image
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&formBody); err != nil {
			log.WithError(err).Errorf("resource/images:createImage() %s : Failed to encode request body as Image", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to create image - JSON marshal error",
				StatusCode: http.StatusBadRequest,
			}
		}
		// validate input format
		if err := validation.ValidateUUIDv4(formBody.ID); err != nil {
			log.WithError(err).Errorf("resource/images:createImage() %s : Invalid image UUID format", message.InvalidInputProtocolViolation)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Invalid image UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		for i, _ := range formBody.FlavorIDs {
			if err := validation.ValidateUUIDv4(formBody.FlavorIDs[i]); err != nil {
				log.Errorf("resource/images:createImage() %s : Invalid flavor UUID format", message.InvalidInputProtocolViolation)
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "Invalid flavor UUID format",
					StatusCode: http.StatusBadRequest,
				}
			}
		}

		cLog := log.WithField("image", formBody)
		if err := db.ImageRepository().Create(&formBody); err != nil {
			switch err {
			case repository.ErrImageAssociationAlreadyExists:
				cLog.WithError(err).Errorf("resource/images:createImage() %s : Image with UUID already exists", message.AppRuntimeErr)
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    fmt.Sprintf("image with UUID %s is already registered", formBody.ID),
					StatusCode: http.StatusConflict,
				}
			case repository.ErrImageAssociationDuplicateFlavor:
				cLog.WithError(err).Errorf("resource/images:createImage() %s : One or more flavor IDs is already associated with this image", message.AppRuntimeErr)
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    fmt.Sprintf("one or more flavor ids in %v is already associated with image %s", formBody.FlavorIDs, formBody.ID),
					StatusCode: http.StatusConflict,
				}
			case repository.ErrImageAssociationFlavorDoesNotExist:
				cLog.WithError(err).Errorf("resource/images:createImage() %s : One or more flavor IDs does not exist", message.AppRuntimeErr)
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    fmt.Sprintf("one or more flavor ids in %v does not point to a registered flavor", formBody.FlavorIDs),
					StatusCode: http.StatusBadRequest,
				}
			case repository.ErrImageAssociationDuplicateImageFlavor:
				cLog.WithError(err).Errorf("resource/images:createImage() %s : Image can only be associated with a single ImageFlavor", message.AppRuntimeErr)
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "image can only be associated with one flavor that has FlavorPart = IMAGE",
					StatusCode: http.StatusConflict,
				}
			default:
				cLog.WithError(err).Errorf("resource/images:createImage() %s : Unexpected error when creating image", message.AppRuntimeErr)
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "Unexpected error when creating image, check input format",
					StatusCode: http.StatusBadRequest,
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err := json.NewEncoder(w).Encode(formBody)
		if err != nil {
			cLog.WithError(err).Errorf("resource/images:createImage() %s :Unexpected error when encoding request back to JSON", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Unexpected error when encoding request back to JSON",
				StatusCode: http.StatusInternalServerError,
			}
		}
		cLog.Debug("resource/images:createImage() Successfully created Image")
		return nil
	}
}

// This method is used to check if the key for an image file is cached.
// If the key is cached, the method you return the key ID.
func getKeyFromCache(imageUUID string) (keycache.Key, error) {
	log.Trace("Entered resource/images:getKeyFromCache()")
	defer log.Trace("Exited resource/images:getKeyFromCache()")
	key, exists := keycache.Get(imageUUID)
	if exists && key.ID != "" && time.Now().Before(key.Expired) {
		return key, nil
	}
	return keycache.Key{}, errors.New("resource/images:getKeyFromCache() key is not cached or expired")
}

// This method is used add the key to cache and map it with the image UUID
func cacheKeyInMemory(imageUUID string, keyID string, key []byte) error {
	log.Trace("Entered resource/images:cacheKeyInMemory()")
	defer log.Trace("Exited resource/images:cacheKeyInMemory()")
	keycache.Store(imageUUID, keycache.Key{ID: keyID, Bytes: key, Created: time.Now(), Expired: time.Now().Add(time.Second * time.Duration(consts.DefaultKeyCacheSeconds))})
	return nil
}

func validateSamlSignature(saml Saml, certPemSlice [][]byte) bool {
	// SAML report expiry validation
	if !time.Now().After(saml.Subject.NotBefore) || !time.Now().Before(saml.Subject.NotOnOrAfter) {
		log.Error("resource/images:validateSamlSignature() SAML report not valid")
		return false
	}

	samlReportCert := strings.ReplaceAll(saml.Signature, "\n", "")

	// check if the signatures from one of the CA certs match the signature from the SAML report
	for _, certPem := range certPemSlice {
		block, _ := pem.Decode(certPem)
		if samlReportCert == base64.StdEncoding.EncodeToString(block.Bytes) {
			return true
		}
	}
	return false
}

func verifySamlSignatureAndCertChain(rootCAPems [][]byte, saml Saml) bool {
	// build the trust root CAs first
	roots := x509.NewCertPool()
	for _, rootPEM := range rootCAPems {
		roots.AppendCertsFromPEM(rootPEM)
	}

	verifyRootCAOpts := x509.VerifyOptions{
		Roots: roots,
	}

	cacerts, err := ioutil.ReadFile(constants.SamlCaCertFilePath)
	if err != nil {
		log.Error("resource/images:verifySamlSignatureAndCertChain() Failed to read SAML ca-certificates")
		log.Tracef("%+v", err)
		return false
	}

	certPemSlice, err := getCertificate(cacerts)
	if err != nil {
		log.Errorf("resource/images:verifySamlSignatureAndCertChain() Error while retrieving certificate")
		return false
	}

	log.Debug("resource/images:verifySamlSignatureAndCertChain() Validating saml signature from HVS")
	isValidated := validateSamlSignature(saml, certPemSlice)
	if !isValidated {
		log.Errorf("resource/images:verifySamlSignatureAndCertChain() SAML signature verification failed")
		return false
	}

	log.Debug("resource/images:verifySamlSignatureAndCertChain() Successfully validated SAML signature")
	for _, certPem := range certPemSlice {
		var cert *x509.Certificate
		var err error
		cert, verifyRootCAOpts.Intermediates, err = crypt.GetCertAndChainFromPem(certPem)
		if err != nil {
			log.Errorf("resource/images:verifySamlSignatureAndCertChain() Error while retrieving certificate and intermediates")
			continue
		}

		if !(cert.IsCA && cert.BasicConstraintsValid) {
			if _, err := cert.Verify(verifyRootCAOpts); err != nil {
				log.Errorf("resource/images:verifySamlSignatureAndCertChain() Error while verifying certificate chain: %s", err.Error())
				continue
			}
		}
		log.Info("resource/images:verifySamlSignatureAndCertChain() SAML certificate chain verification successful")
		return true
	}
	log.Info("resource/images:verifySamlSignatureAndCertChain() SAML certificate chain verification failed")
	return false
}

func getCertificate(signingCertPems interface{}) ([][]byte, error) {
	var certPemSlice [][]byte

	switch signingCertPems.(type) {
	case nil:
		return nil, errors.New("Empty ")
	case [][]byte:
		certPemSlice = signingCertPems.([][]byte)
	case []byte:
		certPemSlice = [][]byte{signingCertPems.([]byte)}
	default:
		log.Errorf("signingCertPems has to be of type []byte or [][]byte")
		return nil, errors.New("signingCertPems has to be of type []byte or [][]byte")

	}
	return certPemSlice, nil
}
