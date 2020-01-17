/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package resource

import (
	"encoding/json"
	"fmt"
	"intel/isecl/workload-service/constants"
	"net/http"
	"strconv"
	"intel/isecl/lib/common/log/message"
	"intel/isecl/lib/common/validation"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// SetReportEndpoints
func SetReportsEndpoints(r *mux.Router, db repository.WlsDatabase) {
	log.Trace("resource/reports:SetReportsEndpoints() Entering")
	defer log.Trace("resource/reports:SetReportsEndpoints() Leaving")
	r.HandleFunc("", errorHandler(requiresPermission(getReport(db), []string{constants.ReportsRetrieve}))).Methods("GET")
	r.HandleFunc("", errorHandler(requiresPermission(createReport(db), []string{constants.ReportsCreate}))).Methods("POST").Headers("Content-Type", "application/json")
	r.HandleFunc("/{id}",
		errorHandler(requiresPermission(deleteReportByID(db), []string{constants.ReportsDelete}))).Methods("DELETE")
	r.HandleFunc("/{badid}", badId)
}

func getReport(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/reports:getReport() Entering")
		defer log.Trace("resource/reports:getReport() Leaving")

		var cLog = log
		filterCriteria := repository.ReportFilter{}
		filterCriteria.Filter = true

		// if no parameters were provided, just return an empty reports array
		if len(r.URL.Query()) == 0 {
			log.Errorf("resource/reports:getReport() %s : Query params missing in request", message.InvalidInputBadParam)
			http.Error(w, "At least one query parameter is required", http.StatusBadRequest)
			return nil
		}

		vmID, ok := r.URL.Query()["vm_id"]
		if ok && len(vmID[0]) >= 1 {
			if err := validation.ValidateUUIDv4(vmID[0]); err != nil {
				log.WithError(err).WithError(err).Errorf("resource/reports:getReport() %s : Invalid VM UUID format", message.InvalidInputProtocolViolation)
				return &endpointError{Message: "Failed to retrieve report", StatusCode: http.StatusBadRequest}
			}
			filterCriteria.VMID = vmID[0]
			cLog = log.WithField("VMUUID", vmID[0])
		}

		reportID, ok := r.URL.Query()["report_id"]
		if ok && len(reportID[0]) >= 1 {
			if err := validation.ValidateUUIDv4(reportID[0]); err != nil {
				log.WithError(err).Errorf("resource/reports:getReport() %s : Invalid report UUID format", message.InvalidInputProtocolViolation)
				log.Tracef("%+v", err)
				return &endpointError{Message: "Failed to retrieve report", StatusCode: http.StatusBadRequest}
			}
			filterCriteria.ReportID = reportID[0]
			cLog = cLog.WithField("ReportID", reportID[0])
		}

		hardwareUUID, ok := r.URL.Query()["hardware_uuid"]
		if ok && len(hardwareUUID[0]) >= 1 {
			if err := validation.ValidateHardwareUUID(hardwareUUID[0]); err != nil {
				log.WithError(err).Errorf("resource/reports:getReport() %s : Invalid hardware UUID format", message.InvalidInputProtocolViolation)
				log.Tracef("%+v", err)
				return &endpointError{Message: "Failed to retrieve report", StatusCode: http.StatusBadRequest}
			}
			filterCriteria.HardwareUUID = hardwareUUID[0]
			cLog = cLog.WithField("HardwareUUID", hardwareUUID[0])
		}

		fromDate, ok := r.URL.Query()["from_date"]
		if ok && len(fromDate[0]) >= 1 {
			if err := validation.ValidateDate(fromDate[0]); err != nil {
				log.WithError(err).Errorf("resource/reports:getReport() %s : Invalid from date format. Expected date format mm-dd-yyyy", message.InvalidInputProtocolViolation)
				log.Tracef("%+v", err)
				return &endpointError{Message: "Failed to retrieve report", StatusCode: http.StatusBadRequest}
			}
			filterCriteria.FromDate = fromDate[0]
			cLog = cLog.WithField("fromDate", fromDate[0])
		}

		toDate, ok := r.URL.Query()["to_date"]
		if ok && len(toDate[0]) >= 1 {
			if err := validation.ValidateDate(toDate[0]); err != nil {
				cLog.WithError(err).Errorf("resource/reports:getReport() %s : Invalid to date format. Expected date format mm-dd-yyyy", message.InvalidInputProtocolViolation)
				log.Tracef("%+v", err)
				return &endpointError{Message: "Failed to retrieve report", StatusCode: http.StatusBadRequest}
			}
			filterCriteria.ToDate = toDate[0]
			cLog = cLog.WithField("toDate", toDate[0])
		}

		latestPerVM, ok := r.URL.Query()["latest_per_vm"]
		if ok && len(latestPerVM[0]) >= 1 {
			boolValue, err := strconv.ParseBool(latestPerVM[0])
			if err != nil {
				cLog.WithError(err).Errorf("resource/reports:getReport() %s : Invalid latest_per_vm boolean value, must be true or false", message.InvalidInputProtocolViolation)
				log.Tracef("%+v", err)
				return &endpointError{Message: "Failed to retrieve report", StatusCode: http.StatusBadRequest}
			}
			filterCriteria.LatestPerVM = boolValue
		} else {
			filterCriteria.LatestPerVM = false
		}

		numOfDays, ok := r.URL.Query()["num_of_days"]
		if ok && len(numOfDays[0]) >= 1 {
			nd, err := strconv.Atoi(numOfDays[0])
			if err != nil {
				cLog.WithError(err).Errorf("resource/reports:getReport() %s : Invalid integer value for num_of_days query parameter", message.InvalidInputProtocolViolation)
				log.Tracef("%+v", err)
				return &endpointError{Message: "Failed to retrieve report", StatusCode: http.StatusBadRequest}
			}
			filterCriteria.NumOfDays = nd
		}

		filter, ok := r.URL.Query()["filter"]
		if ok && len(filter[0]) >= 1 {
			boolValue, err := strconv.ParseBool(filter[0])
			if err != nil {
				cLog.WithError(err).Errorf("resource/reports:getReport() %s : Invalid filter boolean value, must be true or false", message.InvalidInputProtocolViolation)
				log.Tracef("%+v", err)
				return &endpointError{Message: "Failed to retrieve report", StatusCode: http.StatusBadRequest}
			}
			filterCriteria.Filter = boolValue
		}
		cLog.Debugf("HWID: %s|ReportID: %s|VMID: %s|ToDate: %s|FromDate: %s|NumOfDays: %d|Filter: %t|LatestPerVM: %t", filterCriteria.HardwareUUID,
			filterCriteria.ReportID, filterCriteria.VMID, filterCriteria.ToDate, filterCriteria.FromDate, filterCriteria.NumOfDays, filterCriteria.Filter, filterCriteria.LatestPerVM)

		if filterCriteria.HardwareUUID == "" && filterCriteria.ReportID == "" && filterCriteria.VMID == "" && filterCriteria.ToDate == "" && filterCriteria.FromDate == "" && filterCriteria.NumOfDays <= 0 && filterCriteria.Filter {
			cLog.Errorf("resource/reports:getReport() %s : Invalid filter criteria. Allowed filter critierias are vm_id, report_id, hardware_uuid, from_date, to_date, latest_per_vm, nums_of_days and filter = false\n", message.InvalidInputProtocolViolation)
			return &endpointError{Message: "Invalid filter criteria. Allowed filter critierias are vm_id, report_id, hardware_uuid, from_date, to_date, latest_per_vm, nums_of_days and filter = false", StatusCode: http.StatusBadRequest}
		}

		reports, err := db.ReportRepository().RetrieveByFilterCriteria(filterCriteria)
		if err != nil {
			cLog.WithError(err).Errorf("resource/reports:getReport() %s : Failed to retrieve reports", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{Message: "Failed to retrieve reports", StatusCode: http.StatusInternalServerError}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(reports); err != nil {
			cLog.WithError(err).Errorf("resource/reports:getReport() %s : Unexpectedly failed to encode reports to JSON", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{Message: "Failed to retrieve reports - JSON encode failed", StatusCode: http.StatusInternalServerError}
		}
		cLog.Debug("resource/reports:getReport() Successfully retrieved report")
		return nil
	}
}

func createReport(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/reports:createReport() Entering")
		defer log.Trace("resource/reports:createReport() Leaving")

		var vtr model.Report
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&vtr); err != nil {
			log.Errorf("resource/reports:createReport() %s : Report creation failed", message.AppRuntimeErr)
			return &endpointError{
				Message:    "Report creation failed",
				StatusCode: http.StatusBadRequest,
			}
		}

		if err := json.Unmarshal(vtr.Data, &vtr.InstanceTrustReport); err != nil {
			log.Errorf("resource/reports:createReport() %s : Report creation failed", message.AppRuntimeErr)
			return &endpointError{
				Message:    "Report creation failed",
				StatusCode: http.StatusBadRequest,
			}
		}

		// it's almost silly that we unmarshal, then remarshal it to store it back into the database, but at least it provides some validation of the input
		rr := db.ReportRepository()

		// Performance Related:
		// currently, we don't decipher the creation error to see if Creation failed because a collision happened between a primary or unique key.
		// It would be nice to know why the record creation fails, and return the proper http status code.
		// It could be done several ways:
		// - Type assert the error back to PSQL (should be done in the repository layer), and bubble up that information somehow
		// - Manually run a query to see if anything exists with uuid or label (should be done in the repository layer, so we can execute it in a transaction)
		//    - Currently doing this ^
		cLog := log.WithField("report", vtr)
		switch err := rr.Create(&vtr); err {
		case errors.New("resource/reports:createReport() report already exists with UUID"):
			msg := fmt.Sprintf("Report with UUID %s already exists", vtr.Manifest.InstanceInfo.InstanceID)
			cLog.Errorf("resource/reports:createReport() %s : %s", message.InvalidInputBadParam, msg)
			return &endpointError{
				Message:    msg,
				StatusCode: http.StatusConflict,
			}
		case nil:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(vtr); err != nil {
				cLog.WithError(err).Errorf("resource/reports:createReport() %s : Unexpectedly failed to encode Report to JSON", message.AppRuntimeErr)
				log.Tracef("%+v", err)
				return &endpointError{
					Message:    "Failed to create reports - JSON encode failed",
					StatusCode: http.StatusConflict,
				}
			}
			return nil
		default:
			cLog.WithError(err).Errorf("resource/reports:createReport() %s : Unexpected error when creating report", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Unexpected error when creating report, check input format",
				StatusCode: http.StatusBadRequest,
			}
		}
	}
}

func deleteReportByID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		log.Trace("resource/reports:deleteReportByID() Entering")
		defer log.Trace("resource/reports:deleteReportByID() Leaving")

		uuid := mux.Vars(r)["id"]
		// validate UUID
		if err := validation.ValidateUUIDv4(uuid); err != nil {
			log.WithError(err).Errorf("resource/reports:deleteReportByID() %s : Invalid report UUID format: %s", message.InvalidInputProtocolViolation, uuid)
			log.Tracef("%+v", err)
			return &endpointError{Message: "Failed to delete report by UUID", StatusCode: http.StatusBadRequest}
		}
		cLog := log.WithField("uuid", uuid)

		// TODO: Potential dupe check. Shouldn't this be validated by the ValidateUUIDv4 call above?
		if uuid == "" {
			log.Errorf("resource/reports:deleteReportByID() %s : Report id cannot be empty", message.InvalidInputBadParam)
			return &endpointError{
				Message:    "Report id cannot be empty",
				StatusCode: http.StatusBadRequest,
			}
		}
		if err := db.ReportRepository().DeleteByReportID(uuid); err != nil {
			cLog.WithError(err).Errorf("resource/reports:deleteReportByID() %s : Failed to delete Report by UUID", message.AppRuntimeErr)
			log.Tracef("%+v", err)
			return &endpointError{
				Message:    "Report id cannot be empty",
				StatusCode: http.StatusInternalServerError,
			}
		}
		w.WriteHeader(http.StatusNoContent)
		cLog.Debug("resource/reports:deleteReportByID() Successfully deleted Report by UUID")
		return nil
	}
}
