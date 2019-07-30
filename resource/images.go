package resource

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	kms "intel/isecl/lib/kms-client"
	"intel/isecl/workload-service/config"
	consts "intel/isecl/workload-service/constants"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"intel/isecl/lib/common/validation"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// SetImagesEndpoints sets endpoints for /image
func SetImagesEndpoints(r *mux.Router, db repository.WlsDatabase) {
	r.HandleFunc(fmt.Sprintf("/{id:%s}/flavors", uuidv4),
		errorHandler(requiresPermission(getAllAssociatedFlavors(db), []string{consts.AdministratorGroupName}))).Methods("GET")
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
	r.HandleFunc(fmt.Sprintf("/{id:%s}/flavors", uuidv4),
		errorHandler(requiresPermission(retrieveFlavorForImageID(db), []string{consts.AdministratorGroupName, consts.FlavorImageRetrievalGroupName}))).Methods("GET").Queries("flavor_part", "{flavor_part}")
	r.HandleFunc("",
		(errorHandler(requiresPermission(queryImages(db), []string{consts.AdministratorGroupName})))).Methods("GET")
	r.HandleFunc("",
		(errorHandler(requiresPermission(createImage(db), []string{consts.AdministratorGroupName})))).Methods("POST").Headers("Content-Type", "application/json")
	r.HandleFunc("/{badid}", badId)
}

func badId(w http.ResponseWriter, r *http.Request) {
	badid := mux.Vars(r)["badid"]
	log.WithField("uuid", badid).Info("Request made with non compliant UUIDv4")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintf("%s is not uuidv4 compliant", badid)))
}

func missingQueryParameters(params ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		errStr := fmt.Sprintf("Missing query parameters: %v", params)
		log.Debug(errStr)
		http.Error(w, errStr, http.StatusBadRequest)
	}
}

func retrieveFlavorAndKeyForImageID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		// validate UUID format
		if err := validation.ValidateUUIDv4(id); err != nil {
			log.Error("Invalid UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		hwid := mux.Vars(r)["hardware_uuid"]
		// validate hardware UUID
		if err := validation.ValidateHardwareUUID(hwid); err != nil {
			log.Error("Invalid hardware UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		cLog := log.WithFields(log.Fields{
			"imageUUID":    id,
			"hardwareUUID": hwid,
		})
		if hwid == "" {
			cLog.Debug("Missing required parameter hardware_uuid")
			return &endpointError{
				Message:    "Query parameter 'hardware_uuid' cannot be nil",
				StatusCode: http.StatusBadRequest,
			}
		}
		kid, kidPresent := r.URL.Query()["key_id"]
		cLog = cLog.WithField("keyID", kid)
		cLog.Debug("Retrieving Flavor and Key for Image")
		flavor, err := db.ImageRepository().RetrieveAssociatedImageFlavor(id)
		if err != nil {
			cLog.Info("Failed to retrieve Flavor and Key for Image")
			return err
		} else {
			cLog.Info(flavor)
		}
		// Check if flavor keyURL is not empty
		if len(flavor.ImageFlavor.Encryption.KeyURL) > 0 {
			// we have key URL
			// http://10.1.68.21:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer"

			// post HVS with hardwareUUID
			// extract key_id from KeyURL
			cLog = cLog.WithField("keyURL", flavor.ImageFlavor.Encryption.KeyURL)
			cLog.Debug("KeyURL is present")
			keyURL, err := url.Parse(flavor.ImageFlavor.Encryption.KeyURL)
			if err != nil {
				cLog.WithError(err).Error("Flavor KeyURL is malformed")
				return err
			}
			re := regexp.MustCompile("(?i)([0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})")
			keyID := re.FindString(keyURL.Path)
			if !kidPresent || (kidPresent && kid[0] != keyID) {
				criteriaJSON := []byte(fmt.Sprintf(`{"hardware_uuid":"%s"}`, hwid))
				url, err := url.Parse(config.Configuration.HVS.URL)
				if err != nil {
					cLog.WithError(err).Error("Configured HVS URL is malformed: ", err)
					return err
				}
				reports, _ := url.Parse("reports")
				endpoint := url.ResolveReference(reports)
				req, err := http.NewRequest("POST", endpoint.String(), bytes.NewBuffer(criteriaJSON))
				req.SetBasicAuth(config.Configuration.HVS.User, config.Configuration.HVS.Password)
				if err != nil {
					cLog.WithError(err).Error("Failed to instantiate http request to HVS")
					return err
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept", "application/samlassertion+xml")
				client := &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true,
						},
					},
				}
				resp, err := client.Do(req)
				if err != nil {
					cLog.WithError(err).Error("Failed to perform HTTP request to HVS")
					return err
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					text, _ := ioutil.ReadAll(resp.Body)
					errStr := fmt.Sprintf("HVS request failed to retrieve host report (HTTP Status Code: %d)\nMessage: %s", resp.StatusCode, string(text))
					cLog.WithField("statusCode", resp.StatusCode).Info(errStr)
					return &endpointError{
						Message:    errStr,
						StatusCode: http.StatusBadRequest,
					}
				}
				saml, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					cLog.WithError(err).Error("Faield to read HVS response body")
					return err
				}
				cLog.WithField("saml", string(saml)).Debug("Successfully got SAML report from HVS")
				// create insecure client
				kc := &kms.Client{
					BaseURL:    config.Configuration.KMS.URL,
					Username:   config.Configuration.KMS.User,
					Password:   config.Configuration.KMS.Password,
					HTTPClient: client,
				}
				// post to KBS client with saml
				key, err := kc.Key(keyID).Transfer(saml)
				if err != nil {
					cLog.WithError(err).Info("Failed to retrieve key from KMS")
					if kmsErr, ok := err.(*kms.Error); ok {
						return &endpointError{
							Message:    kmsErr.Message,
							StatusCode: kmsErr.StatusCode,
						}
					}
					return err
				}
				cLog.WithField("key", key).Debug("Successfully got key from KMS")
				// got key data
				flavorKey := model.FlavorKey{
					Flavor:    flavor.ImageFlavor,
					Signature: flavor.Signature,
					Key:       key,
				}
				if err := json.NewEncoder(w).Encode(flavorKey); err != nil {
					// marshalling error 500
					cLog.WithError(err).Error("Unexpectedly failed to encode FlavorKey to JSON")
					return err
				}
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				cLog.WithField("flavorKey", flavorKey).Debug("Susccessfully retrieved FlavorKey")
				return nil
			}
		}
		// just return the flavor
		if err := json.NewEncoder(w).Encode(model.FlavorKey{Flavor: flavor.ImageFlavor, Signature: flavor.Signature}); err != nil {
			// marshalling error 500
			cLog.WithError(err).Error("Unexpectedly failed to encode FlavorKey to JSON")
			return err
		}
		return nil
	}
}

func retrieveFlavorForImageID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		// validate UUID
		if err := validation.ValidateUUIDv4(id); err != nil {
			log.Error("Invalid UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		fp := mux.Vars(r)["flavor_part"]
		// validate flavor part
		fpArr := []string{fp}
		if validateInputErr := validation.ValidateStrings(fpArr); validateInputErr != nil {
			log.Error("Invalid flavor part string format")
			return &endpointError{Message: validateInputErr.Error(), StatusCode: http.StatusBadRequest}
		}
		cLog := log.WithFields(log.Fields{
			"imageUUID":  id,
			"flavorPart": fp,
		})
		if fp == "" {
			cLog.Debug("Missing required parameter flavor_part")
			return &endpointError{
				Message:    "Query parameter 'flavor_part' cannot be nil",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog.Debug("Retrieving Flavor for Image")
		flavor, err := db.ImageRepository().RetrieveAssociatedFlavorByFlavorPart(id, fp)
		if err != nil {
			cLog.Info("Failed to retrieve Flavor for Image")
			return err
		}

		// just return the flavor
		if err := json.NewEncoder(w).Encode(*flavor); err != nil {
			// marshalling error 500
			cLog.WithError(err).Error("Unexpectedly failed to encode FlavorKey to JSON")
			return err
		}
		return nil
	}
}

func getAllAssociatedFlavors(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		uuid := mux.Vars(r)["id"]
		// validate UUID format
		if err := validation.ValidateUUIDv4(uuid); err != nil {
			log.Error("Invalid UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		cLog := log.WithField("uuid", uuid)
		flavors, err := db.ImageRepository().RetrieveAssociatedFlavors(uuid)
		if err != nil {
			cLog.WithError(err).Info("Failed to retrieve associated flavors for image defg")
			if err.Error() == "record not found" {
				cLog.Info("No Flavor found for Image")
				json.NewEncoder(w).Encode(flavors)
				return nil
			}
			return err
		}
		if err := json.NewEncoder(w).Encode(flavors); err != nil {
			cLog.WithError(err).Error("Unexpectedly failed to encode list of flavors to JSON")
			return err
		}
		cLog.WithField("flavors", flavors).Debug("Successfully retrieved associated flavors for image")
		return nil
	}
}

func getAssociatedFlavor(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		imageUUID := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(imageUUID); err != nil {
			log.Error("Invalid image UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		flavorUUID := mux.Vars(r)["flavorID"]
		// validate flavor UUID
		if err := validation.ValidateUUIDv4(flavorUUID); err != nil {
			log.Error("Invalid flavor UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		cLog := log.WithFields(log.Fields{
			"imageUUID":  imageUUID,
			"flavorUUID": flavorUUID,
		})
		flavor, err := db.ImageRepository().RetrieveAssociatedFlavor(imageUUID, flavorUUID)
		if err != nil {
			cLog.Info("Failed to retrieve associated flavor for image")
			return err
		}
		cLog = cLog.WithField("flavor", flavor)
		if err := json.NewEncoder(w).Encode(flavor); err != nil {
			cLog.WithError(err).Error("Unexpectedly failed to encode Flavor to JSON")
			return err
		}
		cLog.Debug("Successfully retrieved associated Flavor")
		return nil
	}
}

func putAssociatedFlavor(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		imageUUID := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(imageUUID); err != nil {
			log.Error("Invalid image UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		flavorUUID := mux.Vars(r)["flavorID"]
		// validate flavor UUID
		if err := validation.ValidateUUIDv4(flavorUUID); err != nil {
			log.Error("Invalid flavor UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		cLog := log.WithFields(log.Fields{
			"imageUUID":  imageUUID,
			"flavorUUID": flavorUUID,
		})
		if err := db.ImageRepository().AddAssociatedFlavor(imageUUID, flavorUUID); err != nil {
			cLog.WithError(err).Error("Failed to add new Flavor association")
			return err
		}
		w.WriteHeader(http.StatusCreated)
		cLog.Debug("Successfully added new Flavor association")
		return nil
	}
}

func deleteAssociatedFlavor(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		imageUUID := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(imageUUID); err != nil {
			log.Error("Invalid image UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		flavorUUID := mux.Vars(r)["flavorID"]
		// validate flavor UUID
		if err := validation.ValidateUUIDv4(flavorUUID); err != nil {
			log.Error("Invalid flavor UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		cLog := log.WithFields(log.Fields{
			"imageUUID":  imageUUID,
			"flavorUUID": flavorUUID,
		})
		err := db.ImageRepository().DeleteAssociatedFlavor(imageUUID, flavorUUID)
		if err != nil {
			cLog.Error("Failed to remove Flavor association for Image")
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		cLog.Debug("Successfully removed Flavor association for Image")
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
				log.Error("Invalid filter boolean value, must be true or false")
				return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
			}
			locator.Filter = boolValue
		}

		flavorID, ok := r.URL.Query()["flavor_id"]
		if ok && len(flavorID) >= 1 {
			if err := validation.ValidateUUIDv4(flavorID[0]); err != nil {
				log.Error("Invalid flavor UUID format")
				return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
			}
			locator.FlavorID = flavorID[0]
		}

		imageID, ok := r.URL.Query()["image_id"]
		if ok && len(imageID) >= 1 {
			if err := validation.ValidateUUIDv4(imageID[0]); err != nil {
				log.Error("Invalid image UUID format")
				return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
			}
			locator.ImageID = imageID[0]
		}

		cLog := log.WithFields(log.Fields{
			"image_id":  imageID,
			"flavor_id": flavorID,
			"filter":    filter,
		})

		images, err := db.ImageRepository().RetrieveByFilterCriteria(locator)
		if err != nil {
			cLog.Error("Failed to retrieve Images by filter criteria")
			return err
		}
		if images == nil {
			// coerce to return empty list instead of null
			images = []model.Image{}
		}
		cLog.WithField("images", images)
		if err := json.NewEncoder(w).Encode(images); err != nil {
			cLog.WithError(err).Error("Unexpectedly failed to encode array of Images to JSON")
			return err
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		cLog.Debug("Successfully queried Images by filter criteria")
		return nil
	}
}

func getImageByID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		uuid := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(uuid); err != nil {
			log.Error("Invalid image UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		cLog := log.WithField("uuid", uuid)
		image, err := db.ImageRepository().RetrieveByUUID(uuid)
		if err != nil {
			cLog.WithError(err).Info("Failed to retrieve Image by UUID")
			return err
		}
		cLog = cLog.WithField("image", image)
		if err := json.NewEncoder(w).Encode(image); err != nil {
			cLog.WithError(err).Error("Unexpectedly failed to encode Image to JSON")
			return err
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		cLog.Debug("Successfully retrieved Image by UUID")
		return nil
	}
}

func deleteImageByID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		uuid := mux.Vars(r)["id"]
		// validate image UUID
		if err := validation.ValidateUUIDv4(uuid); err != nil {
			log.Error("Invalid image UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		cLog := log.WithField("uuid", uuid)
		if err := db.ImageRepository().DeleteByUUID(uuid); err != nil {
			cLog.WithError(err).Info("Failed to delete Image by UUID")
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		cLog.Debug("Successfully deleted Image by UUID")
		return nil
	}
}

func createImage(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var formBody model.Image
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&formBody); err != nil {
			log.WithError(err).Info("Failed to encode request body as Image")
			return &endpointError{
				Message:    err.Error(),
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("image", formBody)
		if err := db.ImageRepository().Create(&formBody); err != nil {
			switch err {
			case repository.ErrImageAssociationAlreadyExists:
				cLog.Info("Image with UUID already exists")
				return &endpointError{
					Message:    fmt.Sprintf("image with UUID %s is already registered", formBody.ID),
					StatusCode: http.StatusConflict,
				}
			case repository.ErrImageAssociationDuplicateFlavor:
				cLog.Info("One or more flavor IDs is already associated with this image")
				return &endpointError{
					Message:    fmt.Sprintf("one or more flavor ids in %v is already associated with image %s", formBody.FlavorIDs, formBody.ID),
					StatusCode: http.StatusConflict,
				}
			case repository.ErrImageAssociationFlavorDoesNotExist:
				cLog.Info("One or more flavor IDs does not exist")
				return &endpointError{
					Message:    fmt.Sprintf("one or more flavor ids in %v does not point to a registered flavor", formBody.FlavorIDs),
					StatusCode: http.StatusBadRequest,
				}
			case repository.ErrImageAssociationDuplicateImageFlavor:
				cLog.Info("Image can only be associated with a single ImageFlavor")
				return &endpointError{
					Message:    "image can only be associated with one flavor that has FlavorPart = IMAGE",
					StatusCode: http.StatusConflict,
				}
			default:
				cLog.WithError(err).Error("Unexpected error when creating image")
				return &endpointError{
					Message:    "Unexpected error when creating image, check input format",
					StatusCode: http.StatusBadRequest,
				}
			}
		}
		w.WriteHeader(http.StatusCreated)
		err := json.NewEncoder(w).Encode(formBody)
		if err != nil {
			cLog.WithError(err).Error("Unexpected error when encoding request back to JSON")
			return err
		}
		cLog.Debug("Successfully created Image")
		return nil
	}
}
