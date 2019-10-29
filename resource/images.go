/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package resource

import (
	"crypto/tls"
	"encoding/json"
	"time"

	"encoding/xml"
	"fmt"
	"intel/isecl/lib/common/validation"
	kms "intel/isecl/lib/kms-client"
	consts "intel/isecl/workload-service/constants"
	"intel/isecl/workload-service/keycache"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
	"intel/isecl/workload-service/vsclient"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/pkg/errors"
)

// Saml is used to represent saml report struct
type Saml struct {
	XMLName   xml.Name    `xml:"Assertion"`
	Attribute []Attribute `xml:"AttributeStatement>Attribute"`
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
	r.HandleFunc(fmt.Sprintf("/{id:%s}/flavors", uuidv4),
		(errorHandler(requiresPermission(retrieveFlavorForImageID(db), []string{consts.AdministratorGroupName, consts.FlavorImageRetrievalGroupName})))).Methods("GET").Queries("flavor_part", "{flavor_part}")
	r.HandleFunc(fmt.Sprintf("/{id:%s}/flavors", uuidv4),
		(errorHandler(requiresPermission(getAllAssociatedFlavors(db), []string{consts.AdministratorGroupName})))).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/{id:%s}/flavors/{flavorID:%s}", uuidv4, uuidv4),
		errorHandler(requiresPermission(getAssociatedFlavor(db), []string{consts.AdministratorGroupName}))).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/{id:%s}/flavors/{flavorID:%s}", uuidv4, uuidv4),
		(errorHandler(requiresPermission(putAssociatedFlavor(db), []string{consts.AdministratorGroupName})))).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/{id:%s}/flavors/{flavorID:%s}", uuidv4, uuidv4),
		errorHandler(requiresPermission(deleteAssociatedFlavor(db), []string{consts.AdministratorGroupName}))).Methods("DELETE")
	r.HandleFunc(fmt.Sprintf("/{id:%s}", uuidv4),
		(errorHandler(requiresPermission(getImageByID(db), []string{consts.AdministratorGroupName})))).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/{id:%s}", uuidv4),
		(errorHandler(requiresPermission(deleteImageByID(db), []string{consts.AdministratorGroupName})))).Methods("DELETE")
	r.HandleFunc(fmt.Sprintf("/{id:%s}/flavor-key", uuidv4),
		(errorHandler(requiresPermission(retrieveFlavorAndKeyForImageID(db), []string{consts.AdministratorGroupName, consts.FlavorImageRetrievalGroupName})))).Methods("GET").Queries("hardware_uuid", "{hardware_uuid}")
	r.HandleFunc(fmt.Sprintf("/{id:%s}/flavor-key", uuidv4),
		(missingQueryParameters("hardware_uuid"))).Methods("GET")
	r.HandleFunc("",
		(errorHandler(requiresPermission(queryImages(db), []string{consts.AdministratorGroupName})))).Methods("GET")
	r.HandleFunc("",
		(errorHandler(requiresPermission(createImage(db), []string{consts.AdministratorGroupName})))).Methods("POST").Headers("Content-Type", "application/json")
	r.HandleFunc("/{badid}", badId)
}

func badId(w http.ResponseWriter, r *http.Request) {
	log.Trace("resource/images:badId() Entering")
	defer log.Trace("resource/images:badId() Leaving")
	badid := mux.Vars(r)["badid"]
	log.Errorf("resource/images:badId() Request made with non compliant UUIDv4: %v", badid)

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintf("%s is not uuidv4 compliant", badid)))
}

func missingQueryParameters(params ...string) http.HandlerFunc {
	log.Trace("resource/images:missingQueryParameters() Entering")
	defer log.Trace("resource/images:missingQueryParameters() Leaving")
	return func(w http.ResponseWriter, r *http.Request) {
		errStr := fmt.Sprintf("Missing query parameters: %v", params)
		log.Errorf("resource/images:missingQueryParameters() %s", errStr)
		http.Error(w, errStr, http.StatusBadRequest)
	}
}

func retrieveFlavorAndKeyForImageID(db repository.WlsDatabase) endpointHandler {
	log.Trace("resource/images:retrieveFlavorAndKeyForImageID() Entering")
	defer log.Trace("resource/images:retrieveFlavorAndKeyForImageID() Leaving")

	return func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		// validate UUID format
		if err := validation.ValidateUUIDv4(id); err != nil {
			log.Error("resource/images:retrieveFlavorAndKeyForImageID() Invalid UUID format")
			log.Tracef("%+v", err)
			return &endpointError{
				Message: "Failed to retrieve Flavor/Key - Invalid UUID",
				StatusCode: http.StatusBadRequest,
			}
		}
		hwid := mux.Vars(r)["hardware_uuid"]
		// validate hardware UUID
		if err := validation.ValidateHardwareUUID(hwid); err != nil {
			log.Error("resource/images:retrieveFlavorAndKeyForImageID() Invalid hardware UUID format")
			log.Tracef("%+v", err)
			return &endpointError{
				Message: "Failed to retrieve Flavor/Key - Invalid hardware uuid",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("imageUUID", id).WithField("hardwareUUID", hwid)

		// TODO: Potential dupe check. Shouldn't this be validated by the ValidateHardwareUUID call above?
		if hwid == "" {
			cLog.Debug("resource/images:retrieveFlavorAndKeyForImageID() Missing required parameter hardware_uuid")
			return &endpointError{
				Message:    "Query parameter 'hardware_uuid' cannot be nil",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog.Debug("resource/images:retrieveFlavorAndKeyForImageID() Retrieving Flavor and Key for Image")
		flavor, err := db.ImageRepository().RetrieveAssociatedImageFlavor(id)
		if err != nil {
			cLog.WithError(err).Error("resource/images:retrieveFlavorAndKeyForImageID() Failed to retrieve Flavor and Key for Image")
			return &endpointError{
				Message:    "Failed to retrieve Flavo/Key for Image - Backend Error",
				StatusCode: http.StatusInternalServerError,
			}
		}
		// Check if flavor keyURL is not empty
		if len(flavor.ImageFlavor.Encryption.KeyURL) > 0 {
			// we have key URL
			// http://10.1.68.21:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer"
			// post HVS with hardwareUUID
			// extract key_id from KeyURL
			cLog = cLog.WithField("keyURL", flavor.ImageFlavor.Encryption.KeyURL)
			cLog.Debug("resource/images:retrieveFlavorAndKeyForImageID() KeyURL is present")
			keyURL, err := url.Parse(flavor.ImageFlavor.Encryption.KeyURL)
			if err != nil {
				cLog.WithError(err).Error("resource/images:retrieveFlavorAndKeyForImageID() Flavor KeyURL is malformed")
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "Failed to retrieve Flavor/Key for Image - Flavor KeyURL is malformed",
					StatusCode: http.StatusBadRequest,
				}
			}
			re := regexp.MustCompile("(?i)([0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})")
			keyID := re.FindString(keyURL.Path)

			// retrieve host SAML report from HVS
			saml, err := vsclient.CreateSAMLReport(hwid)
			if err != nil {
				cLog.WithError(err).Error("resource/images:retrieveFlavorAndKeyForImageID() Failed to read HVS response body")
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "Failed to retrieve Flavor/Key for Image - Failed to read HVS response",
					StatusCode: http.StatusInternalServerError,
				}
			}

			// validate the response from HVS
			if err = validation.ValidateXMLString(string(saml)); err != nil {
				cLog.WithError(err).Error("resource/images:retrieveFlavorAndKeyForImageID() HVS response validation failed")
				return &endpointError{
					Message:    "Failed to retrieve Flavor/Key for Image - Invalid SAML report format received from HVS",
					StatusCode: http.StatusInternalServerError,
				}
			}
			var samlStruct Saml
			cLog.WithField("saml", string(saml)).Debug("resource/images:retrieveFlavorAndKeyForImageID() Successfully got SAML report from HVS")
			err = xml.Unmarshal(saml, &samlStruct)
			if err != nil {
				cLog.WithError(err).Error("resource/images:retrieveFlavorAndKeyForImageID() Failed to unmarshal host SAML report")
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "Failed to retrieve Flavor and Key - Failed to unmarshal host SAML report",
					StatusCode: http.StatusInternalServerError,
				}
			}

			var key []byte
			for i := 0; i < len(samlStruct.Attribute); i++ {
				if samlStruct.Attribute[i].Name == "TRUST_OVERALL" && samlStruct.Attribute[i].AttributeValue == "true" {
					// check if the key is cached and retrieve it
					// try to obtain the key from the cache. If the key is not found in the cache,
					// then it will return and error. In this case, we ignore it and pro

					var cachedKeyID string
					cachedKey, err := getKeyFromCache(id)
					if err == nil {
						cachedKeyID = cachedKey.ID
						cLog.Debugf("resource/images:retrieveFlavorAndKeyForImageID() Retrieved Key from in-memory cache. key ID:%s, imageuuid: %s", cachedKeyID, id)
					}
					// check if the key cached is same as the one in the flavor
					if cachedKeyID != "" && cachedKeyID == keyID {
						key = cachedKey.Bytes
					} else {
						// create insecure client
						client := &http.Client{
							Transport: &http.Transport{
								TLSClientConfig: &tls.Config{
									InsecureSkipVerify: true,
								},
							},
						}
						kc := &kms.Client{
							BaseURL:    keyURL.String(),
							HTTPClient: client,
						}
						// post to KBS client with saml
						key, err = kc.Key(keyID).Transfer(saml)
						if err != nil {
							cLog.WithError(err).Error("resource/images:retrieveFlavorAndKeyForImageID() Failed to retrieve key from KMS")
							if kmsErr, ok := err.(*kms.Error); ok {
								return &endpointError{
									Message:    "Failed to retrieve key - " +  kmsErr.Message,
									StatusCode: kmsErr.StatusCode,
								}
							}
							return &endpointError{
								Message:    "Failed to retrieve Key from KMS",
								StatusCode: http.StatusInternalServerError,
							}
						}
						cLog.WithField("key", key).Debug("resource/images:retrieveFlavorAndKeyForImageID() Successfully got key from KMS")
						cacheKeyInMemory(id, keyID, key)
					}
				}
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
				cLog.WithError(err).Error("resource/images:retrieveFlavorAndKeyForImageID() Unexpectedly failed to encode FlavorKey to JSON")
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "Failed to encode FlavorKey to JSON - Failure marshalling JSON response",
					StatusCode: http.StatusInternalServerError,
				}
			}
			cLog.WithField("flavorKey", flavorKey).Debug("resource/images:retrieveFlavorAndKeyForImageID() Successfully retrieved FlavorKey")
			return nil
		}
		// just return the flavor
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(model.FlavorKey{Flavor: flavor.ImageFlavor, Signature: flavor.Signature}); err != nil {
			// marshalling error 500
			cLog.WithError(err).Error("resource/images:retrieveFlavorAndKeyForImageID() Unexpectedly failed to encode FlavorKey to JSON")
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve FlavorKey - Failure marshalling JSON response",
				StatusCode: http.StatusInternalServerError,
			}
		}
		cLog.Debug("resource/images:retrieveFlavorAndKeyForImageID() Successfully retrieved Flavor and Key")
		return nil
	}
}

func retrieveFlavorForImageID(db repository.WlsDatabase) endpointHandler {
	log.Trace("resource/images:retrieveFlavorForImageID() Entering")
	defer log.Trace("resource/images:retrieveFlavorForImageID() Leaving")

	return func(w http.ResponseWriter, r *http.Request) error {
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
			log.WithError(validateInputErr).Error("resource/images:retrieveFlavorForImageID() Invalid flavor part string format")
			return &endpointError{
				Message: "Failed to retrieve flavor - Invalid flavor part string format",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("imageUUID", id).WithField("flavorPart", fp)

		if fp == "" {
			cLog.Debug("resource/images:retrieveFlavorForImageID() Missing required parameter flavor_part")
			return &endpointError{
				Message:    "Failed to retrieve flavor - Query parameter 'flavor_part' cannot be nil",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog.Debug("resource/images:retrieveFlavorForImageID() Retrieving Flavor for Image")
		flavor, err := db.ImageRepository().RetrieveAssociatedFlavorByFlavorPart(id, fp)
		if err != nil {
			cLog.WithError(err).Error("resource/images:retrieveFlavorForImageID() Failed to retrieve Flavor for Image")
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve flavor - backend error",
				StatusCode: http.StatusInternalServerError,
			}
		}

		// just return the flavor
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(*flavor); err != nil {
			// marshalling error 500
			cLog.WithError(err).Error("resource/images:retrieveFlavorForImageID() Unexpectedly failed to encode FlavorKey to JSON")
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
	log.Trace("resource/images:getAllAssociatedFlavors() Entering")
	defer log.Trace("resource/images:getAllAssociatedFlavors() Leaving")

	return func(w http.ResponseWriter, r *http.Request) error {
		uuid := mux.Vars(r)["id"]
		// validate UUID format
		if err := validation.ValidateUUIDv4(uuid); err != nil {
			log.WithError(err).Error("resource/images:getAllAssociatedFlavors() Invalid UUID format")
			log.Tracef("%+v", err)
			return &endpointError{
				Message: "Failed to retrieve associated flavors - Invalid UUID format", 
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("uuid", uuid)
		flavors, err := db.ImageRepository().RetrieveAssociatedFlavors(uuid)
		if err != nil {
			if err.Error() == "record not found" {
				cLog.Info("resource/images:getAllAssociatedFlavors() No Flavor found for Image")
				log.Tracef("%+v", err)
				json.NewEncoder(w).Encode(flavors)
				return nil
			} else {
				cLog.WithError(err).Error("resource/images:getAllAssociatedFlavors() Failed to retrieve associated flavors for image defg")
				log.Tracef("%+v", err)
				return &endpointError{
					Message: "Failed to retrieve associated flavors - backend error",
					StatusCode: http.StatusInternalServerError,
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(flavors); err != nil {
			cLog.WithError(err).Error("resource/images:getAllAssociatedFlavors() Unexpectedly failed to encode list of flavors to JSON")
			log.Tracef("%+v", err)
			return &endpointError{
				Message: "Failed to retrieve associated flavors - JSON marshal response failure",
				StatusCode: http.StatusInternalServerError,
			}
		}
		cLog.WithField("flavors", flavors).Debug("resource/images:getAllAssociatedFlavors() Successfully retrieved associated flavors for image")
		return nil
	}
}

func getAssociatedFlavor(db repository.WlsDatabase) endpointHandler {
	log.Trace("resource/images:getAssociatedFlavor() Entering")
	defer log.Trace("resource/images:getAssociatedFlavor() Leaving")
	return func(w http.ResponseWriter, r *http.Request) error {
		imageUUID := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(imageUUID); err != nil {
			log.WithError(err).Error("resource/images:getAssociatedFlavor() Invalid image UUID format")
			log.Tracef("%+v", err)
			return &endpointError{
				Message: "Failed to retrieve flavor - invalid image UUID format", 
				StatusCode: http.StatusBadRequest,
			}
		}
		flavorUUID := mux.Vars(r)["flavorID"]
		// validate flavor UUID
		if err := validation.ValidateUUIDv4(flavorUUID); err != nil {
			log.WithError(err).Error("resource/images:getAssociatedFlavor() Invalid flavor UUID format")
			log.Tracef("%+v", err)
			return &endpointError{
				Message: "Failed to retrieve associated flavor for image - Invalid image UUID format",
				 StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("imageUUID", imageUUID).WithField("flavorUUID", flavorUUID)

		flavor, err := db.ImageRepository().RetrieveAssociatedFlavor(imageUUID, flavorUUID)
		if err != nil {
			cLog.WithError(err).Error("resource/images:getAssociatedFlavor() Failed to retrieve associated flavor for image")
			log.Tracef("%+v", err)
			return &endpointError{
				Message: "No flavor associated with given image UUID",
				StatusCode: http.StatusNotFound,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		cLog = cLog.WithField("flavor", flavor)
		if err := json.NewEncoder(w).Encode(flavor); err != nil {
			cLog.WithError(err).Error("resource/images:getAssociatedFlavor() Unexpectedly failed to encode Flavor to JSON")
			return &endpointError{
				Message: "Failed to retrieve associated flavor for image - JSON response encode marshal failure",
				StatusCode: http.StatusInternalServerError,
			}
		}
		cLog.Debug("resource/images:getAssociatedFlavor() Successfully retrieved associated Flavor")
		return nil
	}
}

func putAssociatedFlavor(db repository.WlsDatabase) endpointHandler {
	log.Trace("resource/images:putAssociatedFlavor() Entering")
	defer log.Trace("resource/images:putAssociatedFlavor() Leaving")

	return func(w http.ResponseWriter, r *http.Request) error {
		imageUUID := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(imageUUID); err != nil {
			log.WithError(err).Error("resource/images:putAssociatedFlavor() Invalid image UUID format")
			log.Tracef("%+v", err)
			return &endpointError{
				Message: "Failed to create image/flavor association - invalid image UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		flavorUUID := mux.Vars(r)["flavorID"]
		// validate flavor UUID
		if err := validation.ValidateUUIDv4(flavorUUID); err != nil {
			log.WithError(err).Error("resource/images:putAssociatedFlavor() Invalid flavor UUID format")
			log.Tracef("%+v", err)
			return &endpointError{
				Message: "Failed to create image/flavor association - invalid flavor UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("imageUUID", imageUUID).WithField("flavorUUID", flavorUUID)

		if err := db.ImageRepository().AddAssociatedFlavor(imageUUID, flavorUUID); err != nil {
			cLog.WithError(err).Error("resource/images:putAssociatedFlavor() Failed to add new Flavor association")
			log.Tracef("%+v", err)
			return &endpointError{
				Message: "Failed to create image/flavor association - invalid image UUID format - Backend error",
				StatusCode: http.StatusInternalServerError,
			}
		}
		w.WriteHeader(http.StatusCreated)
		cLog.Debug("resource/images:putAssociatedFlavor() Successfully added new Flavor association")
		return nil
	}
}

func deleteAssociatedFlavor(db repository.WlsDatabase) endpointHandler {
	log.Trace("resource/images:deleteAssociatedFlavor() Entering")
	defer log.Trace("resource/images:deleteAssociatedFlavor() Leaving")

	return func(w http.ResponseWriter, r *http.Request) error {
		imageUUID := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(imageUUID); err != nil {
			log.WithError(err).Error("resource/images:deleteAssociatedFlavor() Invalid image UUID format")
			log.Tracef("%+v", err)
			return &endpointError{
				Message: "Failed to delete image/flavor association - invalid image UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		flavorUUID := mux.Vars(r)["flavorID"]
		// validate flavor UUID
		if err := validation.ValidateUUIDv4(flavorUUID); err != nil {
			log.WithError(err).Error("resource/images:deleteAssociatedFlavor() Invalid flavor UUID format")
			log.Tracef("%+v", err)
			return &endpointError{
				Message: "Failed to delete image/flavor association - invalid flavor UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("imageUUID", imageUUID).WithField("flavorUUID", flavorUUID)
		err := db.ImageRepository().DeleteAssociatedFlavor(imageUUID, flavorUUID)
		if err != nil {
			cLog.WithError(err).Error("resource/images:deleteAssociatedFlavor() Failed to remove Flavor association for Image")
			log.Tracef("%+v", err)
			return &endpointError{
				Message: "Failed to delete image/flavor association - Backend error",
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
	log.Trace("resource/images:queryImages() Entering")
	defer log.Trace("resource/images:queryImages() Leaving")

	return func(w http.ResponseWriter, r *http.Request) error {
		locator := repository.ImageFilter{}
		locator.Filter = true // default to 'filter' to true

		if len(r.URL.Query()) == 0 {
			http.Error(w, "At least one query parameter is required", http.StatusBadRequest)
			return nil
		}

		filter, ok := r.URL.Query()["filter"]
		if ok && len(filter) >= 1 {
			boolValue, err := strconv.ParseBool(filter[0])
			if err != nil {
				log.WithError(err).Error("resource/images:queryImages() Invalid filter boolean value, must be true or false")
				return &endpointError{
					Message: "Failed to retrieve image - Invalid filter boolean value, must be true or false",
					StatusCode: http.StatusBadRequest,
				}
			}
			locator.Filter = boolValue
		}

		flavorID, ok := r.URL.Query()["flavor_id"]
		if ok && len(flavorID) >= 1 {
			if err := validation.ValidateUUIDv4(flavorID[0]); err != nil {
				log.WithError(err).Error("resource/images:queryImages() Invalid flavor UUID format")
				return &endpointError{
					Message:  "Failed to retrieve image - Invalid flavor UUID format",
					StatusCode: http.StatusBadRequest,
				}
			}
			locator.FlavorID = flavorID[0]
		}

		imageID, ok := r.URL.Query()["image_id"]
		if ok && len(imageID) >= 1 {
			if err := validation.ValidateUUIDv4(imageID[0]); err != nil {
				log.WithError(err).Error("resource/images:queryImages() Invalid image UUID format")
				return &endpointError{
					Message: "Failed to retrieve image - Invalid image UUID format",
					StatusCode: http.StatusBadRequest,
				}
			}
			locator.ImageID = imageID[0]
		}

		cLog := log.WithField("image_id", imageID).WithField("flavor_id", flavorID).WithField("filter", filter)

		if locator.FlavorID == "" && locator.ImageID == "" && locator.Filter {
			log.Error("Invalid filter criteria. Allowed filter critierias are image_id, flavor_id and filter = false\n")
			return &endpointError{
				Message: "Failed to retrieve image - Invalid filter criteria. Allowed filter critierias are image_id, flavor_id and filter",
				StatusCode: http.StatusBadRequest,
			}
		}

		images, err := db.ImageRepository().RetrieveByFilterCriteria(locator)
		if err != nil {
			cLog.WithError(err).Error("resource/images:queryImages() Failed to retrieve Images by filter criteria")
			return &endpointError{
				Message: "Failed to retrieve image - Failed to retrieve Images by filter criteria",
				StatusCode: http.StatusInternalServerError,
			}
		}
		if images == nil {
			// coerce to return empty list instead of null
			images = []model.Image{}
		}
		w.Header().Set("Content-Type", "application/json")
		cLog.WithField("images", images)
		if err := json.NewEncoder(w).Encode(images); err != nil {
			cLog.WithError(err).Error("resource/images:queryImages() Unexpectedly failed to encode array of Images to JSON")
			return &endpointError{
				Message: "Failed to retrieve image - JSON response encode marshal failure",
				StatusCode: http.StatusInternalServerError,
			}
		}
		cLog.Debug("resource/images:queryImages() Successfully queried Images by filter criteria")
		return nil
	}
}

func getImageByID(db repository.WlsDatabase) endpointHandler {
	log.Trace("resource/images:getImageByID() Entering")
	defer log.Trace("resource/images:getImageByID() Leaving")

	return func(w http.ResponseWriter, r *http.Request) error {
		uuid := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(uuid); err != nil {
			log.WithError(err).Error("resource/images:getImageByID() Invalid image UUID format")
			return &endpointError{Message: "Failed to retrieve image - Invalid image UUID format",
				 StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("uuid", uuid)
		image, err := db.ImageRepository().RetrieveByUUID(uuid)
		if err != nil {
			cLog.WithError(err).Error("resource/images:getImageByID() Failed to retrieve Image by UUID")
			return &endpointError{
				Message: "No image found for given UUID",
				StatusCode: http.StatusNotFound,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		cLog = cLog.WithField("image", image)
		if err := json.NewEncoder(w).Encode(image); err != nil {
			cLog.WithError(err).Error("resource/images:getImageByID() Unexpectedly failed to encode Image to JSON")
			return &endpointError{
				Message: "Failed to retrieve image - JSON response encode marshal failure",
				StatusCode: http.StatusInternalServerError,
			}
		}
		cLog.Debug("resource/images:getImageByID() Successfully retrieved Image by UUID")
		return nil
	}
}

func deleteImageByID(db repository.WlsDatabase) endpointHandler {
	log.Trace("resource/images:deleteImageByID() Entering")
	defer log.Trace("resource/images:deleteImageByID() Leaving")

	return func(w http.ResponseWriter, r *http.Request) error {
		uuid := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(uuid); err != nil {
			log.WithError(err).Error("resource/images:deleteImageByID() Invalid image UUID format")
			return &endpointError{
				Message: "Failed to delete image - invalid image UUID", 
			StatusCode: http.StatusBadRequest,
		}
		}
		cLog := log.WithField("uuid", uuid)
		if err := db.ImageRepository().DeleteByUUID(uuid); err != nil {
			cLog.WithError(err).Error("resource/images:deleteImageByID() Failed to delete Image by UUID")
			return &endpointError{
				Message: "Failed to delete image - backend error",
				StatusCode: http.StatusInternalServerError,
			}
		}
		w.WriteHeader(http.StatusNoContent)
		cLog.Debug("resource/images:deleteImageByID() Successfully deleted Image by UUID")
		return nil
	}
}

func createImage(db repository.WlsDatabase) endpointHandler {
	log.Trace("resource/images:createImage() Entering")
	defer log.Trace("resource/images:createImage() Leaving")

	return func(w http.ResponseWriter, r *http.Request) error {
		var formBody model.Image
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&formBody); err != nil {
			log.WithError(err).Error("resource/images:createImage() Failed to encode request body as Image")
			return &endpointError{
				Message:    "Failed to create image - JSON marshal error",
				StatusCode: http.StatusBadRequest,
			}
		}
		// validate input format
		if err := validation.ValidateUUIDv4(formBody.ID); err != nil {
			log.WithError(err).Error("resource/images:createImage() Invalid image UUID format")
			return &endpointError{
				Message: "Invalid image UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		for i, _ := range formBody.FlavorIDs {
			if err := validation.ValidateUUIDv4(formBody.FlavorIDs[i]); err != nil {
				log.Error("resource/images:createImage() Invalid flavor UUID format")
				return &endpointError{
					Message: "Invalid flavor UUID format",
					StatusCode: http.StatusBadRequest,
				}
			}
		}

		cLog := log.WithField("image", formBody)
		if err := db.ImageRepository().Create(&formBody); err != nil {
			switch err {
			case repository.ErrImageAssociationAlreadyExists:
				cLog.WithError(err).Error("resource/images:createImage() Image with UUID already exists")
				return &endpointError{
					Message:    fmt.Sprintf("image with UUID %s is already registered", formBody.ID),
					StatusCode: http.StatusConflict,
				}
			case repository.ErrImageAssociationDuplicateFlavor:
				cLog.WithError(err).Error("resource/images:createImage() One or more flavor IDs is already associated with this image")
				return &endpointError{
					Message:    fmt.Sprintf("one or more flavor ids in %v is already associated with image %s", formBody.FlavorIDs, formBody.ID),
					StatusCode: http.StatusConflict,
				}
			case repository.ErrImageAssociationFlavorDoesNotExist:
				cLog.WithError(err).Error("resource/images:createImage() One or more flavor IDs does not exist")
				return &endpointError{
					Message:    fmt.Sprintf("one or more flavor ids in %v does not point to a registered flavor", formBody.FlavorIDs),
					StatusCode: http.StatusBadRequest,
				}
			case repository.ErrImageAssociationDuplicateImageFlavor:
				cLog.WithError(err).Error("resource/images:createImage() Image can only be associated with a single ImageFlavor")
				return &endpointError{
					Message:    "image can only be associated with one flavor that has FlavorPart = IMAGE",
					StatusCode: http.StatusConflict,
				}
			default:
				cLog.WithError(err).Error("resource/images:createImage() Unexpected error when creating image")
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
			cLog.WithError(err).Error("resource/images:createImage() Unexpected error when encoding request back to JSON")
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
	keycache.Store(imageUUID, keycache.Key{keyID, key, time.Now(), time.Now().Add(time.Second * time.Duration(consts.DefaultKeyCacheSeconds))})
	return nil
}
