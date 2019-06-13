package repository

import (
	"intel/isecl/workload-service/model"
)

// ReportRepository defines an interface that provides persistence operations for a Flavor.
// It defines High Level CRUD operations that could be implemented by any database or persistence layer (such as postgres)
// The CRUD operations are logically grouped, but not defined to any single interface, so that FlavorRepository may customize them to its own needs, with
// Stronger typing rather than cast everything from an interface{}
type ReportRepository interface {
	// C
	Create(r *model.Report) error
	// R
	RetrieveByFilterCriteria(filter ReportFilter) ([]model.Report, error)
	// D
	DeleteByReportID(uuid string) error
}

// ReportFilter struct defines all the filter criterias to query the reports table
type ReportFilter struct {
	VMID         string `json:"vm_id,omitempty"`
	ReportID     string `json:"report_id,omitempty"`
	HardwareUUID string `json:"hardware_uuid,omitempty"`
	LatestPerVM  string `json:"latest_per_vm,omitempty"`
	ToDate       string `json:"to_date,omitempty"`
	FromDate     string `json:"from_date,omitempty"`
	NumOfDays    int    `json:"no_of_days,omitempty"`
	Filter       bool   `json:"filter,omitempty"`
}
