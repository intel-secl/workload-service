package resource

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"errors"
	"strconv"

	"intel/isecl/workload-service/config"
	"intel/isecl/lib/middleware/logger"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// SetReportEndpoints
func SetReportsEndpoints(r *mux.Router, db *gorm.DB) {
	logger := logger.NewLogger(config.LogWriter, "WLS - ", log.Ldate|log.Ltime)
	r.HandleFunc("", logger(getReport(db))).Methods("GET")
	r.HandleFunc("", logger(createReport(db))).Methods("POST").Headers("Content-Type", "application/json").Headers("Accept", "application/json")
	r.HandleFunc("/{id}", logger(deleteReportByID(db))).Methods("DELETE")
}

func getReport(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		
		locator := repository.ReportLocator{}
		vmID, ok := r.URL.Query()["vm_id"]
		if ok && len(vmID) >= 1 {
			locator.VmID = vmID[0]
		}
		
		reportID, ok := r.URL.Query()["report_id"]
		if ok && len(reportID) >= 1 {
			locator.ReportID = reportID[0]
		}
		
		hardwareUUID, ok := r.URL.Query()["hardware_uuid"]
		if ok && len(hardwareUUID) >= 1 {
			locator.HardwareUUID = hardwareUUID[0]
		}
		
		fromDate, ok := r.URL.Query()["from_date"]
		if ok && len(fromDate) >= 1 {
			locator.FromDate = fromDate[0]
		}
		
		toDate, ok := r.URL.Query()["to_date"]
		if ok && len(toDate) >= 1 {
			locator.ToDate = toDate[0]
		}

		latestPerVM, ok := r.URL.Query()["latest_per_vm"]
		if ok && len(latestPerVM) >= 1 {
			locator.LatestPerVM = latestPerVM[0]
		}

		numOfDays, ok := r.URL.Query()["num_of_days"]
		if ok && len(numOfDays) >= 1 {
			nd, err := strconv.Atoi(numOfDays[0])
			if err == nil{
				locator.NumOfDays = nd
			}
		}
		
		reports, err := repository.GetReportRepository(db).RetrieveByFilterCriteria(locator)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		if err := json.NewEncoder(w).Encode(reports); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
		}
	}
}

func createReport(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var vtr model.VMTrustReport
        
		if err := json.NewDecoder(r.Body).Decode(&vtr); err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		// it's almost silly that we unmarshal, then remarshal it to store it back into the database, but at least it provides some validation of the input
		rr := repository.GetReportRepository(db)

		// Performance Related:
		// currently, we don't decipher the creation error to see if Creation failed because a collision happened between a primary or unique key.
		// It would be nice to know why the record creation fails, and return the proper http status code.
		// It could be done several ways:
		// - Type assert the error back to PSQL (should be done in the repository layer), and bubble up that information somehow
		// - Manually run a query to see if anything exists with uuid or label (should be done in the repository layer, so we can execute it in a transaction)
		//    - Currently doing this ^
		switch err := rr.Create(&vtr); err {
		case errors.New("report already exists with UUID"):
			http.Error(w, fmt.Sprintf("Report with UUID %s already exists", vtr.Manifest.VmInfo.VmID), http.StatusConflict)
		case nil:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(vtr); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} 
			
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)  
		}
	}
}

func deleteReportByID(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := mux.Vars(r)["id"]
		if uuid == ""{
			http.Error(w, "Report id empty", http.StatusBadRequest)
		}
		err := repository.GetReportRepository(db).DeleteByReportID(uuid)
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
