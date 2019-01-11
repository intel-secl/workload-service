package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"intel/isecl/lib/common/logger"
	httpLogger "intel/isecl/lib/middleware/logger"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"

	"github.com/gorilla/mux"
)

// SetReportEndpoints
func SetReportsEndpoints(r *mux.Router, db repository.WlsDatabase) {
	logger := httpLogger.NewLogger(logger.Info)
	r.HandleFunc("", logger(errorHandler(getReport(db)))).Methods("GET")
	r.HandleFunc("", logger(errorHandler(createReport(db)))).Methods("POST").Headers("Content-Type", "application/json").Headers("Accept", "application/json")
	r.HandleFunc("/{id:(?i:[0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})}", logger(errorHandler(deleteReportByID(db)))).Methods("DELETE")
}

func getReport(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		filter := repository.ReportFilter{}
		vmID, ok := r.URL.Query()["vm_id"]
		if ok && len(vmID) >= 1 {
			filter.VMID = vmID[0]
		}

		reportID, ok := r.URL.Query()["report_id"]
		if ok && len(reportID) >= 1 {
			filter.ReportID = reportID[0]
		}

		hardwareUUID, ok := r.URL.Query()["hardware_uuid"]
		if ok && len(hardwareUUID) >= 1 {
			filter.HardwareUUID = hardwareUUID[0]
		}

		fromDate, ok := r.URL.Query()["from_date"]
		if ok && len(fromDate) >= 1 {
			filter.FromDate = fromDate[0]
		}

		toDate, ok := r.URL.Query()["to_date"]
		if ok && len(toDate) >= 1 {
			filter.ToDate = toDate[0]
		}

		latestPerVM, ok := r.URL.Query()["latest_per_vm"]
		if ok && len(latestPerVM) >= 1 {
			filter.LatestPerVM = latestPerVM[0]
		}

		numOfDays, ok := r.URL.Query()["num_of_days"]
		if ok && len(numOfDays) >= 1 {
			nd, err := strconv.Atoi(numOfDays[0])
			if err == nil {
				filter.NumOfDays = nd
			}
		}

		reports, err := db.ReportRepository().RetrieveByFilterCriteria(filter)
		if err != nil {
			logger.Info.Println("Failed to retrive reports: ", err)
			return err
		}

		if err := json.NewEncoder(w).Encode(reports); err != nil {
			logger.Error.Println("Unexpectedly failed to encode reports to JSON", err)
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
		if err := json.NewDecoder(r.Body).Decode(&vtr); err != nil {
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
		switch err := rr.Create(&vtr); err {
		case errors.New("report already exists with UUID"):
			msg := fmt.Sprintf("Report with UUID %s already exists", vtr.Manifest.VmInfo.VmID)
			logger.Info.Println(msg)
			return &endpointError{
				Message:    msg,
				StatusCode: http.StatusConflict,
			}
		case nil:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(vtr); err != nil {
				logger.Error.Println(err)
				return err
			}
			return nil
		default:
			logger.Error.Println(err)
			return err
		}
	}
}

func deleteReportByID(db repository.WlsDatabase) endpointHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		uuid := mux.Vars(r)["id"]
		if uuid == "" {
			return &endpointError{
				Message:    "Report id cannot be empty",
				StatusCode: http.StatusBadRequest,
			}
		}
		if err := db.ReportRepository().DeleteByReportID(uuid); err != nil {
			logger.Info.Println("Failed to delete Report by id: ", err)
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		logger.Info.Println("Successfully deleted Report with id: ", uuid)
		return nil
	}
}
