/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package resource

import (
	"fmt"
	"intel/isecl/workload-service/repository"
	"intel/isecl/workload-service/version"
	"net/http"

	"github.com/gorilla/mux"
)

// SetVersionEndpoints installs route handler for GET /version
func SetVersionEndpoints(r *mux.Router, db repository.WlsDatabase) {
	log.Trace("Entered resource/version:SetVersionEndpoints()")
	defer log.Trace("Exited resource/version:SetVersionEndpoints()")
	r.HandleFunc("", getVersion).Methods("GET")
}

// GetVersion handles GET /version
func getVersion(w http.ResponseWriter, r *http.Request) {
	log.Trace("resource/version:getVersion() Entering")
	defer log.Trace("resource/version:getVersion() Leaving")
	w.WriteHeader(http.StatusOK)
	log.Debugf("resource/version:getVersion() WLS Version: %s CommitHash: %s", version.Version, version.GitHash)
	w.Write([]byte(fmt.Sprintf("%s-%s", version.Version, version.GitHash)))
}
