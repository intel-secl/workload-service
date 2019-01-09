package resource

import (
	"io/ioutil"
	"regexp"
	"net/url"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"intel/isecl/lib/middleware/logger"
	"intel/isecl/workload-service/config"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
	"intel/isecl/lib/kms-client"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// SetImagesEndpoints sets endpoints for /image
func SetImagesEndpoints(r *mux.Router, db repository.WlsDatabase) {
	logger := logger.NewLogger(config.LogWriter, "WLS - ", log.Ldate|log.Ltime)
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}/flavors", logger(errorHandler(getAllAssociatedFlavors(db)))).Methods("GET")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}/flavors/{flavorID:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", 
		logger(errorHandler(getAssociatedFlavor(db)))).Methods("GET")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}/flavors/{flavorID:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", 
		logger(errorHandler(putAssociatedFlavor(db)))).Methods("PUT")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}/flavors/{flavorID:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", 
		logger(deleteAssociatedFlavor(db))).Methods("DELETE")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}",
		logger(errorHandler(getImageByID(db)))).Methods("GET")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", 
		logger(errorHandler(deleteImageByID(db)))).Methods("DELETE")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}/flavor-key", 
		logger(errorHandler(retrieveFlavorAndKeyForImageID(db)))).Methods("GET").Queries("hardware_uuid", "{hardware_uuid}")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}/flavor-key", logger(missingQueryParameters("hardware_uuid"))).Methods("GET")
	r.HandleFunc("", 
		logger(errorHandler(queryImages(db)))).Methods("GET")
	r.HandleFunc("", 
		logger(errorHandler(createImage(db)))).Methods("POST").Headers("Content-Type", "application/json")
}

func missingQueryParameters(params... string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		errStr := fmt.Sprintf("Missing query parameters: %v", params)
		http.Error(w, errStr, http.StatusBadRequest)
	}
}

func retrieveFlavorAndKeyForImageID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := mux.Vars(r)["id"]
		hwid := mux.Vars(r)["hardware_uuid"]
		if hwid == "" {
			return &endpointError {
				Message: "Query parameter 'hardware_uuid' cannot be nil",
				StatusCode: http.StatusBadRequest,
			}
		}
		kid, kidPresent := r.URL.Query()["key_id"]
		flavor, err := db.ImageRepository().RetrieveAssociatedImageFlavor(id)
		if err != nil {
			return err
		} 
		// Check if flavor keyURL is not empty
		if len(flavor.Image.Encryption.KeyURL) > 0 {
			// we have key URL
			// http://10.1.68.21:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer"

			// post HVS with hardwareUUID
			// extract key_id from KeyURL
			keyURL, err := url.Parse(flavor.Image.Encryption.KeyURL)
			if err != nil {
				return err
			}
			re := regexp.MustCompile("(?i)([0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})")
			keyID := re.FindString(keyURL.Path)
			if !kidPresent || (kidPresent && kid[0] != keyID){
				criteriaJSON := []byte(fmt.Sprintf(`{"hardware_uuid":"%s"}`, hwid))
				url, err := url.Parse(config.Configuration.HVS.URL)
				if err != nil {
					return err
				}
				reports, err := url.Parse("reports")
				if err != nil {
					return err
				}
				endpoint := url.ResolveReference(reports)
				req, err := http.NewRequest("POST", endpoint.String(), bytes.NewBuffer(criteriaJSON))
				req.SetBasicAuth(config.Configuration.HVS.User, config.Configuration.HVS.Password)
				if err != nil {
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
					// bad response from HVS from the http side
					return err
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					text, _ := ioutil.ReadAll(resp.Body)
					errStr := fmt.Sprintf("HVS request failed to retrieve host report (HTTP Status Code: %d)\nMessage: %s", resp.StatusCode, string(text))
					return &endpointError{
						Message: errStr,
						StatusCode: http.StatusBadRequest,
					}
				}
				saml, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				// create insecure client
				kc := &kms.Client{
					BaseURL: config.Configuration.KMS.URL,
					Username: config.Configuration.KMS.User,
					Password: config.Configuration.KMS.Password,
					HTTPClient: client,
				}
				// post to KBS client with saml
				key, err := kc.Key(keyID).Transfer(saml)
				if err != nil {
					if kmsErr, ok := err.(*kms.Error); ok {
						return &endpointError {
							Message: kmsErr.Message, 
							StatusCode: kmsErr.StatusCode,
						}
					}
					return err
				}
				// got key data
				flavorKey := model.FlavorKey{
					Flavor: *flavor,
					Key: key,
				}
				if err := json.NewEncoder(w).Encode(flavorKey); err != nil {
					// marshalling error 500
					return err
				}
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/json")
				return nil
			}
		}
		// just return the flavor
		if err := json.NewEncoder(w).Encode(model.FlavorKey{Flavor: *flavor}); err != nil {
			// marshalling error 500
			return err
		}
		return nil
	}
}


func getAllAssociatedFlavors(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		uuid := mux.Vars(r)["id"]
		flavors, err := db.ImageRepository().RetrieveAssociatedFlavors(uuid)
		if err != nil {
			return err
		}
		if err := json.NewEncoder(w).Encode(flavors); err != nil {
			return err
		}
		return nil
	}
}

func getAssociatedFlavor(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		imageUUID := mux.Vars(r)["id"]
		flavorUUID := mux.Vars(r)["flavorID"]
		flavor, err := db.ImageRepository().RetrieveAssociatedFlavor(imageUUID, flavorUUID)
		if err != nil {
			return err
		}
		if err := json.NewEncoder(w).Encode(flavor); err != nil {
			return err
		}
		return nil
	}
}

func putAssociatedFlavor(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		imageUUID := mux.Vars(r)["id"]
		flavorUUID := mux.Vars(r)["flavorID"]
		if err := db.ImageRepository().AddAssociatedFlavor(imageUUID, flavorUUID); err != nil {
			return err
		}
		w.WriteHeader(http.StatusCreated)
		return nil
	}
}

func deleteAssociatedFlavor(db repository.WlsDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		imageUUID := mux.Vars(r)["id"]
		flavorUUID := mux.Vars(r)["flavorID"]
		err := db.ImageRepository().DeleteAssociatedFlavor(imageUUID, flavorUUID)
		if err != nil {
			var code int
			if gorm.IsRecordNotFoundError(err) {
				code = http.StatusNotFound
			} else {
				code = http.StatusInternalServerError
			}
			http.Error(w, err.Error(), code)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func queryImages(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		locator := repository.ImageFilter{}
		flavorID, ok := r.URL.Query()["flavor_id"]
		if ok && len(flavorID) >= 1 {
			locator.FlavorID = flavorID[0]
		}

		images, err := db.ImageRepository().RetrieveByFilterCriteria(locator)
		if err != nil {
			return err
		}
		if len(images) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return nil
		}
		if err := json.NewEncoder(w).Encode(images); err != nil {
			return err
		} 
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		return nil
	}
}

func getImageByID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		uuid := mux.Vars(r)["id"]
		image, err := db.ImageRepository().RetrieveByUUID(uuid)
		if err != nil {
			return err
		} 
		if err := json.NewEncoder(w).Encode(image); err != nil {
			return err
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		return nil
	}
}

func deleteImageByID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		uuid := mux.Vars(r)["id"]
		if err := db.ImageRepository().DeleteByUUID(uuid); err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}

func createImage(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var formBody model.Image
		if err := json.NewDecoder(r.Body).Decode(&formBody); err != nil {
			return &endpointError {
				Message: err.Error(),
				StatusCode: http.StatusBadRequest,
			}
		}
		if err := db.ImageRepository().Create(&formBody); err != nil {
			switch err {
			case repository.ErrImageAssociationAlreadyExists:
				return &endpointError {
					Message: fmt.Sprintf("image with UUID %s is already registered", formBody.ID), 
					StatusCode: http.StatusConflict,
				}
			case repository.ErrImageAssociationDuplicateFlavor:
				return &endpointError {
					Message: fmt.Sprintf("one or more flavor ids in %v is already associated with image %s", formBody.FlavorIDs, formBody.ID), 
					StatusCode: http.StatusConflict,
				}
			case repository.ErrImageAssociationFlavorDoesNotExist:
				return &endpointError {
					Message: fmt.Sprintf("one or more flavor ids in %v does not point to a registered flavor", formBody.FlavorIDs), 
					StatusCode: http.StatusBadRequest,
				}
			case repository.ErrImageAssociationDuplicateImageFlavor:
				return &endpointError {
					Message: "image can only be associated with one flavor that has FlavorPart = IMAGE", 
					StatusCode: http.StatusConflict,
				}
			default:
				return err
			}
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(formBody)
		return nil
	}
}
