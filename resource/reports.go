package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
	"intel/isecl/lib/common/validation"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// SetReportEndpoints
func SetReportsEndpoints(r *mux.Router, db repository.WlsDatabase) {
	r.HandleFunc("", (errorHandler(getReport(db)))).Methods("GET")
	r.HandleFunc("", (errorHandler(createReport(db)))).Methods("POST").Headers("Content-Type", "application/json")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", (errorHandler(deleteReportByID(db)))).Methods("DELETE")
	r.HandleFunc("/{badid}", badId)
}

func getReport(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		filterCriteria := repository.ReportFilter{}

		// if no parameters were provided, just return an empty reports array
		if len(r.URL.Query()) == 0 {
			http.Error(w, "At least one query parameter is required", http.StatusBadRequest)
			return nil
		}

		vmID, ok := r.URL.Query()["vm_id"]
		if ok && len(vmID) >= 1 {
			if err := validation.ValidateUUIDv4(vmID[0]); err != nil {
				log.Error("Invalid VM UUID format")
				return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
			}
			filterCriteria.VMID = vmID[0]
		}

		reportID, ok := r.URL.Query()["report_id"]
		if ok && len(reportID) >= 1 {
			if err := validation.ValidateUUIDv4(reportID[0]); err != nil {
				log.Error("Invalid report UUID format")
				return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
			}
			filterCriteria.ReportID = reportID[0]
		}

		hardwareUUID, ok := r.URL.Query()["hardware_uuid"]
		if ok && len(hardwareUUID) >= 1 {
			if err := validation.ValidateHardwareUUID(hardwareUUID[0]); err != nil {
				log.Error("Invalid hardware UUID format")
				return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
			}
			filterCriteria.HardwareUUID = hardwareUUID[0]
		}

		fromDate, ok := r.URL.Query()["from_date"]
		if ok && len(fromDate) >= 1 {
			if err := validation.ValidateDate(fromDate[0]); err != nil {
				log.Error("Invalid from date format. Expected date format mm-dd-yyyy")
				return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
			}	
			filterCriteria.FromDate = fromDate[0]
		}

		toDate, ok := r.URL.Query()["to_date"]
		if ok && len(toDate) >= 1 {
			if err := validation.ValidateDate(toDate[0]); err != nil {
				log.Error("Invalid from to format. Expected date format mm-dd-yyyy")
				return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
			}
			filterCriteria.ToDate = toDate[0]
		}

		latestPerVM, ok := r.URL.Query()["latest_per_vm"]
		if ok && len(latestPerVM) >= 1 {
			boolValue, err := strconv.ParseBool(latestPerVM[0])
			if err != nil {
				log.Error("Invalid latest_per_vm boolean value, must be true or false")
				return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
			}
			filterCriteria.LatestPerVM = boolValue
		}

		numOfDays, ok := r.URL.Query()["num_of_days"]
		if ok && len(numOfDays) >= 1 {
			nd, err := strconv.Atoi(numOfDays[0])
			if err != nil {
				log.Error("Invalid integer value for num_of_days query parameter")
				return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
			}
			filterCriteria.NumOfDays = nd
		}

		filter, ok := r.URL.Query()["filter"]
		if ok && len(filter) >= 1 {
			boolValue, err := strconv.ParseBool(filter[0])
			if err != nil {
				log.Error("Invalid filter boolean value, must be true or false")
				return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
			}
			filterCriteria.Filter = boolValue
		}

		reports, err := db.ReportRepository().RetrieveByFilterCriteria(filterCriteria)
		if err != nil {
			log.WithError(err).Info("Failed to retrieve reports")
			return err
		}

		if err := json.NewEncoder(w).Encode(reports); err != nil {
			log.WithError(err).Error("Unexpectedly failed to encode reports to JSON")
			return err
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		return nil
	}
}

func createReport(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var vtr model.Report
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&vtr); err != nil {
			return &endpointError{
				Message:    err.Error(),
				StatusCode: http.StatusBadRequest,
			}
		}

		if err := json.Unmarshal(vtr.Data, &vtr.InstanceTrustReport); err != nil {
			return &endpointError{
				Message:    err.Error(),
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
		case errors.New("report already exists with UUID"):
			msg := fmt.Sprintf("Report with UUID %s already exists", vtr.Manifest.InstanceInfo.InstanceID)
			cLog.Info(msg)
			return &endpointError{
				Message:    msg,
				StatusCode: http.StatusConflict,
			}
		case nil:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(vtr); err != nil {
				cLog.WithError(err).Error("Unexpectedly failed to encode Report to JSON")
				return err
			}
			return nil
		default:
			cLog.WithError(err).Error("Unexpected error when creating report")
			return &endpointError{
				Message:    "Unexpected error when creating report, check input format",
				StatusCode: http.StatusBadRequest,
			}
		}
	}
}

func deleteReportByID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		uuid := mux.Vars(r)["id"]
		// validate UUID
		if err := validation.ValidateUUIDv4(uuid); err != nil {
			log.Error("Invalid report UUID format")
			return &endpointError{Message: err.Error(), StatusCode: http.StatusBadRequest}
		}
		cLog := log.WithField("uuid", uuid)
		if uuid == "" {
			return &endpointError{
				Message:    "Report id cannot be empty",
				StatusCode: http.StatusBadRequest,
			}
		}
		if err := db.ReportRepository().DeleteByReportID(uuid); err != nil {
			cLog.WithError(err).Info("Failed to delete Report by UUID")
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		cLog.Debug("Successfully deleted Report by UUID")
		return nil
	}
}
