/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package resource

import (
	"encoding/json"
	"intel/isecl/lib/common/v4/log/message"
	"intel/isecl/lib/common/v4/validation"
	consts "intel/isecl/workload-service/v4/constants"
	"intel/isecl/workload-service/v4/model"
	"intel/isecl/workload-service/v4/repository"
	"net/http"

	"github.com/gorilla/mux"
)

// SetKeysEndpoints sets endpoints for /keys
func SetKeysEndpoints(r *mux.Router, db repository.WlsDatabase) {
	log.Trace("resource/keys:SetKeysEndpoints() Entering")
	defer log.Trace("resource/keys:SetKeysEndpoints() Leaving")
	r.HandleFunc("",
		(errorHandler(requiresPermission(retrieveKey(db), []string{consts.KeysCreate})))).Methods("POST").Headers("Content-Type", "application/json")
}

func retrieveKey(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/keys:retrieveKeyEntering()")
		defer log.Trace("resource/keys:retrieveKey() Leaving")

		var formBody model.RequestKey
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&formBody); err != nil {
			log.WithError(err).Errorf("resource/keys:retrieveKey() %s : Failed to encode request body as Key", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Failed to retrieve key - JSON marshal error",
				StatusCode: http.StatusBadRequest,
			}
		}
		// validate input format
		hwid := formBody.HwId
		if err := validation.ValidateHardwareUUID(hwid); err != nil {
			log.WithError(err).Errorf("resource/keys:retrieveKey() %s : Invalid Hardware UUID format", message.InvalidInputProtocolViolation)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Invalid hardware UUID format",
				StatusCode: http.StatusBadRequest,
			}
		}
		cLog := log.WithField("hardwareUUID", hwid)

		cLog.Debug("resource/keys:retrievendKey() Retrieving  Key")

		keyUrl := formBody.KeyUrl
		// Check if flavor keyUrl is not empty
		if len(keyUrl) > 0 {
			key, err := transfer_key(false, hwid, keyUrl, "")
			if err != nil {
				cLog.WithError(err).Error("resource/keys:retrieveKey() Error while retrieving key")
				return err
			}

			// got key data
			returnKey := model.ReturnKey{
				Key: key,
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(returnKey); err != nil {
				// marshalling error 500
				cLog.WithError(err).Errorf("resource/keys:retrieveKey() %s : Unexpectedly failed to encode returnKey to JSON", message.AppRuntimeErr)
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "Failed to retrieve Key - Failure marshalling JSON response",
					StatusCode: http.StatusInternalServerError,
				}
			}
			cLog.Info("resource/keys:retreiveKey() Successfully retrieved Key")
			return nil
		}
		return nil
	}
}
