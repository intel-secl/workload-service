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
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}/flavors", logger(getAllAssociatedFlavors(db))).Methods("GET")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}/flavors/{flavorID:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", 
		logger(getAssociatedFlavor(db))).Methods("GET")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}/flavors/{flavorID:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", 
		logger(putAssociatedFlavor(db))).Methods("PUT")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}/flavors/{flavorID:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", 
		logger(deleteAssociatedFlavor(db))).Methods("DELETE")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}/flavors/{flavorID:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", nil).Methods("DELETE")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", logger(getImageByID(db))).Methods("GET")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", logger(deleteImageByID(db))).Methods("DELETE")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}/flavor-key", logger(retrieveFlavorAndKeyForImageID(db))).Methods("GET").Queries("hardware_uuid", "{hardware_uuid}")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}/flavor-key", logger(missingQueryParameters("hardware_uuid"))).Methods("GET")
	r.HandleFunc("", logger(queryImages(db))).Methods("GET")
	r.HandleFunc("", logger(createImage(db))).Methods("POST").Headers("Content-Type", "application/json")
}

func missingQueryParameters(params... string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		errStr := fmt.Sprintf("Missing query parameters: %v", params)
		http.Error(w, errStr, http.StatusBadRequest)
	}
}

func retrieveFlavorAndKeyForImageID(db repository.WlsDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		hwid := mux.Vars(r)["hardware_uuid"]
		if hwid == "" {
			http.Error(w, "Query parameter 'hardware_uuid' cannot be nil", http.StatusBadRequest)
			return
		}
		kid, kidPresent := r.URL.Query()["key_id"]
		flavor, err := db.ImageRepository().RetrieveAssociatedImageFlavor(id)
		if err != nil {
			var code int
			if gorm.IsRecordNotFoundError(err) {
				code = http.StatusNotFound
			} else {
				code = http.StatusInternalServerError
			}
			http.Error(w, err.Error(), code)
		} else {
			// Check if flavor keyURL is not empty
			if len(flavor.Image.Encryption.KeyURL) > 0 {
				// we have key URL
				// http://10.1.68.21:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer"

				// post HVS with hardwareUUID
				// extract key_id from KeyURL
				keyURL, err := url.Parse(flavor.Image.Encryption.KeyURL)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError) 
					return
				}
				re := regexp.MustCompile("(?i)([0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})")
				keyID := re.FindString(keyURL.Path)
				if !kidPresent || (kidPresent && kid[0] != keyID){
					criteriaJSON := []byte(fmt.Sprintf(`{"hardware_uuid":"%s"}`, hwid))
					url, err := url.Parse(config.Configuration.HVS.URL)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					reports, err := url.Parse("reports")
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					endpoint := url.ResolveReference(reports)
					req, err := http.NewRequest("POST", endpoint.String(), bytes.NewBuffer(criteriaJSON))
					req.SetBasicAuth(config.Configuration.HVS.User, config.Configuration.HVS.Password)
					if err != nil {
						// http error internal 500
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
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
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					defer resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						text, _ := ioutil.ReadAll(resp.Body)
						errStr := fmt.Sprintf("HVS request failed to retrieve host report (HTTP Status Code: %d)\nMessage: %s", resp.StatusCode, string(text))
						http.Error(w, errStr, http.StatusBadRequest)
						return
					}
					saml, _ := ioutil.ReadAll(resp.Body)

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
							http.Error(w, kmsErr.Error(), kmsErr.StatusCode)
							return
						}
						http.Error(w, err.Error(), http.StatusInternalServerError) 
						return
					}
					// got key data
					flavorKey := model.FlavorKey{
						Flavor: *flavor,
						Key: key,
					}
					if err := json.NewEncoder(w).Encode(flavorKey); err != nil {
						// marshalling error 500
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", "application/json")
					return
				}
			}
			if err := json.NewEncoder(w).Encode(model.FlavorKey{Flavor: *flavor}); err != nil {
				// marshalling error 500
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}


func getAllAssociatedFlavors(db repository.WlsDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := mux.Vars(r)["id"]
		flavors, err := db.ImageRepository().RetrieveAssociatedFlavors(uuid)
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
		json.NewEncoder(w).Encode(flavors)
	}
}

func getAssociatedFlavor(db repository.WlsDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		imageUUID := mux.Vars(r)["id"]
		flavorUUID := mux.Vars(r)["flavorID"]
		flavor, err := db.ImageRepository().RetrieveAssociatedFlavor(imageUUID, flavorUUID)
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
		json.NewEncoder(w).Encode(flavor)
	}
}

func putAssociatedFlavor(db repository.WlsDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		imageUUID := mux.Vars(r)["id"]
		flavorUUID := mux.Vars(r)["flavorID"]
		err := db.ImageRepository().AddAssociatedFlavor(imageUUID, flavorUUID)
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
		w.WriteHeader(http.StatusCreated)
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

func queryImages(db repository.WlsDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locator := repository.ImageFilter{}
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
				http.Error(w, fmt.Sprintf("image with UUID %s is already registered", formBody.ID), http.StatusConflict)
			case repository.ErrImageAssociationDuplicateFlavor:
				http.Error(w, fmt.Sprintf("one or more flavor ids in %v is already associated with image %s", formBody.FlavorIDs, formBody.ID), http.StatusConflict)
			case repository.ErrImageAssociationFlavorDoesNotExist:
				http.Error(w, fmt.Sprintf("one or more flavor ids in %v does not point to a registered flavor", formBody.FlavorIDs), http.StatusBadRequest)
			case repository.ErrImageAssociationDuplicateImageFlavor:
				http.Error(w, "image can only be associated with one flavor that has FlavorPart = IMAGE", http.StatusConflict)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(formBody)
	}
}
