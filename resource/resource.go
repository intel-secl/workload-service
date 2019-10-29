/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package resource

import (
	"fmt"
	"github.com/pkg/errors"
	"intel/isecl/lib/common/auth"
	"intel/isecl/lib/common/context"
	commLog "intel/isecl/lib/common/log"
	ct "intel/isecl/lib/common/types/aas"
	consts "intel/isecl/workload-service/constants"
	"intel/isecl/workload-service/repository"

	"net/http"

	"github.com/jinzhu/gorm"

	"github.com/gorilla/mux"
)

var log = commLog.GetDefaultLogger()
var seclog = commLog.GetSecurityLogger()

const uuidv4 = "(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})"

// endpointSetter is a function that takes a Gorilla Mux Subrouter, and an instance of a WlsDatabase connection,
// and allows the end user to set and handle any API endpoints on that upaht
type endpointSetter func(r *mux.Router, db repository.WlsDatabase)

// endpointError is a custom error type that lets the thrower specify an http status code
type endpointError struct {
	Message    string
	StatusCode int
}

type privilegeError struct {
	StatusCode int
	Message    string
}

func (e privilegeError) Error() string {
	log.Trace("resource/resource:Error() Entering")
	defer log.Trace("resource/resource:Error() Leaving")
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}

func (e endpointError) Error() string {
	log.Trace("resource/resource:Error() Entering")
	defer log.Trace("resource/resource:Error() Leaving")
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}

func requiresPermission(eh endpointHandler, roleNames []string) endpointHandler {
	log.Trace("resource/resource:requiresPermission() Entering")
	defer log.Trace("resource/resource:requiresPermission() Leaving")
	return func(w http.ResponseWriter, r *http.Request) error {
		privileges, err := context.GetUserRoles(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Could not get user roles from http context"))
			seclog.Debugf("resource/resource:requiresPermission() Roles: %v | Context: %v", roleNames, r.Context())
			return errors.Wrap(err, "resource/resource:requiresPermission() Could not get user roles from http context")
		}
		reqRoles := make([]ct.RoleInfo, len(roleNames))
		for i, role := range roleNames {
			reqRoles[i] = ct.RoleInfo{Service: consts.ServiceName, Name: role}
		}

		seclog.Debugf("resource/resource:requiresPermission() Req Roles: %v", reqRoles)
		_, foundRole := auth.ValidatePermissionAndGetRoleContext(privileges, reqRoles,
			true)
		if !foundRole {
			w.WriteHeader(http.StatusUnauthorized)
			seclog.Errorf("resource/resource:requiresPermission() Insufficient privileges to access %s", r.RequestURI)
			return &privilegeError{Message: "Insufficient privileges to access " + r.RequestURI, StatusCode: http.StatusUnauthorized}
		}
		return eh(w, r)
	}
}

// endpointHandler is the same as http.ResponseHandler, but returns an error that can be handled by a generic
// middleware handler
type endpointHandler func(w http.ResponseWriter, r *http.Request) error

func errorHandler(eh endpointHandler) http.HandlerFunc {
	log.Trace("resource/resource:errorHandler() Entering")
	defer log.Trace("resource/resource:errorHandler() Leaving")
	return func(w http.ResponseWriter, r *http.Request) {
		if err := eh(w, r); err != nil {
			if gorm.IsRecordNotFoundError(err) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			switch t := err.(type) {
			case *endpointError:
				http.Error(w, t.Message, t.StatusCode)
			case privilegeError:
				http.Error(w, t.Message, t.StatusCode)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}
