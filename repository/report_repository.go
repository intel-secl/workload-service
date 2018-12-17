package repository

import (
	"errors"
	"time"
	"fmt"
	
	"intel/isecl/workload-service/model"

	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	
)


// ReportRepository defines an interface that provides persistence operations for a Flavor.
// It defines High Level CRUD operations that could be implemented by any database or persistence layer (such as postgres)
// The CRUD operations are logically grouped, but not defined to any single interface, so that FlavorRepository may customize them to its own needs, with
// Stronger typing rather than cast everything from an interface{}
type ReportRepository interface {
	// C
	Create(r *model.VMTrustReport) error
	// R
	RetrieveByFilterCriteria(locator ReportLocator) ([]model.Report, error)
	// D
	DeleteByReportID(uuid string) error
}

type ReportLocator struct {
	VmID  string `json:"vm_id, omitempty"`
	ReportID string `json:"report_id, omitempty"`
	HardwareUUID string `json:"hardware_uuid, omitempty"`
	LatestPerVM string `json:"latest_per_vm, omitempty"`
	ToDate string `json:"to_date, omitempty"`
	FromDate string `json:"from_date, omitempty"`
	NumOfDays int `json:"no_of_days, omitempty"`
}

type reportRepo struct {
	db *gorm.DB
}

const dateString string = "2006-01-02 15:04:05.999999-07"

func parseTime(strTime string) time.Time{
	t, err := time.Parse(dateString, strTime)
	if err != nil {
		fmt.Println(err)
	}
	return t
}

func getReportModels (reportEntities []reportEntity) ([]model.Report, error){
	ids := make([]model.Report, len(reportEntities))
	
	for i, v := range reportEntities {
		ids[i] = model.Report{ID: v.ID, Saml: v.Saml, VmID: v.VmID, TrustReport: v.TrustReport}
	}
	fmt.Println(len(ids))
	return ids, nil
}


func (rp *reportRepo) RetrieveByFilterCriteria(locator ReportLocator) ([]model.Report, error) {
	db := rp.db
	var reportEntities []reportEntity

	vmID := ""
	reportID := ""
	hardwareUUID := ""
	var toDate time.Time
	var fromDate time.Time
	latestPerVM := "true"
	
	if len(locator.VmID) > 0 {
		vmID = locator.VmID
	}
	
	if len(locator.ReportID) > 0 {
		reportID = locator.ReportID
	}
	
	if len(locator.HardwareUUID) > 0 {
		hardwareUUID = locator.HardwareUUID
	}

	if len(locator.ToDate) > 0 {
		toDate = parseTime(locator.ToDate)
	}

	if len(locator.FromDate) > 0 {
		fromDate = parseTime(locator.FromDate)
	}
	
	if  locator.NumOfDays > 0 {
		toDate = parseTime(time.Now().Format(dateString))
		fromDate = parseTime(toDate.AddDate( 0, 0, -(locator.NumOfDays)).Format(dateString))
	}
	
	if len(locator.LatestPerVM) > 0{
		latestPerVM = locator.LatestPerVM
	}

	//Only fetch the report since reportid is unique across the table
	if (len(reportID) > 0){
		db.Where("id = ?", reportID).Find(&reportEntities)
		return getReportModels(reportEntities)
	}
	fmt.Println("todate")
	fmt.Println(toDate)
	fmt.Println("todate")
	if (latestPerVM == "true" && toDate.IsZero() && fromDate.IsZero()){
		return findLatestReports(vmID, reportID, hardwareUUID, db)
	}else {
		return findReports(vmID, reportID, hardwareUUID, toDate, fromDate, latestPerVM, db)
	}
	
}

func findLatestReports(vmID string, reportID string, hardwareUUID string, db *gorm.DB) ([]model.Report, error){
	
	var reportEntities []reportEntity
	//Only fetch latest record for given a vm_id
	if (len(vmID) > 0){
		db.Where("vm_id = ?", vmID).Order("created_at desc").First(&reportEntities)
		return getReportModels(reportEntities)
	}
	
	if (len(hardwareUUID) > 0){
		db.Joins("LEFT JOIN report_entities as r2 ON report_entities.vm_id = r2.vm_id AND report_entities.created_at > r2.created_at").Where("r2.vm_id is null and report_entities.trust_report->'vm_manifest'->'vm_info'->>'host_hardware_uuid' = ?", hardwareUUID).Find(&reportEntities)
	    return getReportModels(reportEntities)
	}
	return getReportModels(reportEntities)
}

func findReports(vmID string, reportID string, hardwareUUID string, toDate time.Time, fromDate time.Time, latestPerVM string, db *gorm.DB) ([]model.Report, error){
	var reportEntities []reportEntity
	

	if (!fromDate.IsZero() && !toDate.IsZero()){
		    
		if (len(vmID) > 0){
			if (latestPerVM == "true"){
				db.Where("vm_id = ? and created_at > ? and created_at < ?", vmID, fromDate, toDate).Order("created_at desc").First(&reportEntities)
				return getReportModels(reportEntities)
			}else{
				db.Where("vm_id = ? and created_at > ? and created_at < ?", vmID, fromDate, toDate).Order("created_at desc").Find(&reportEntities)
				return getReportModels(reportEntities)
			}
			
		}
				
		if (len(hardwareUUID) > 0){
			if (latestPerVM == "true"){
				db.Joins("LEFT JOIN report_entities as r2 ON report_entities.vm_id = r2.vm_id AND report_entities.created_at > r2.created_at").Where("r2.vm_id is null AND report_entities.trust_report->'vm_manifest'->'vm_info'->>'host_hardware_uuid' = ? AND report_entities.created_at > ? and report_entities.created_at < ?", hardwareUUID, fromDate, toDate).Find(&reportEntities)
				return getReportModels(reportEntities)
			}else{
				db.Order("created_at desc").Where("trust_report->'vm_manifest'->'vm_info'->>'host_hardware_uuid' = ? AND created_at > ? and created_at < ?", hardwareUUID, fromDate, toDate).Find(&reportEntities)
				return getReportModels(reportEntities)
			}
		}
// If only either num of days or from_date and to_date is the given filter criteria
		if (latestPerVM == "true"){
			db.Joins("LEFT JOIN report_entities as r2 ON report_entities.vm_id = r2.vm_id AND report_entities.created_at > r2.created_at").Where("r2.vm_id is null AND report_entities.created_at > ? and report_entities.created_at < ?", fromDate, toDate).Find(&reportEntities)
			fmt.Println("only num of days")
			return getReportModels(reportEntities)
		}else{
			db.Order("created_at desc").Where("created_at > ? and created_at < ?", fromDate, toDate).Find(&reportEntities)
			return getReportModels(reportEntities)
		}
	}

	if (!fromDate.IsZero()){
		
		if (len(vmID) > 0){
			if (latestPerVM == "true"){
				db.Where("vm_id = ? and created_at > ? ", vmID, fromDate).Order("created_at desc").First(&reportEntities)
				return getReportModels(reportEntities)
			}else{
				db.Where("vm_id = ? and created_at > ? ", vmID, fromDate).Order("created_at desc").Find(&reportEntities)
				return getReportModels(reportEntities)
			}
		}
				
		if (len(hardwareUUID) > 0){
			fmt.Println("I am heere")
			if (latestPerVM == "true"){
				db.Table("report_entities").Joins("LEFT JOIN report_entities as r2 ON report_entities.vm_id = r2.vm_id AND report_entities.created_at > r2.created_at").Where("r2.vm_id is null AND report_entities.trust_report->'vm_manifest'->'vm_info'->>'host_hardware_uuid' = ? AND created_at > ?", hardwareUUID, fromDate).Find(&reportEntities)
				return getReportModels(reportEntities)
			}else{
				db.Order("created_at desc").Where("trust_report->'vm_manifest'->'vm_info'->>'host_hardware_uuid' = ? and created_at > ? ", hardwareUUID, fromDate).Find(&reportEntities)
				return getReportModels(reportEntities)
			}
		}
		
		// If only from_date is the given filter criteria
		if (latestPerVM == "true"){
			fmt.Println("from date")
			db.Joins("LEFT JOIN report_entities as r2 ON report_entities.vm_id = r2.vm_id AND report_entities.created_at > r2.created_at").Where("r2.vm_id is null AND report_entities.created_at > ?", fromDate).Find(&reportEntities)
			return getReportModels(reportEntities)
		}else{
			db.Order("created_at desc").Where("created_at > ?", fromDate).Find(&reportEntities)
			return getReportModels(reportEntities)
		}


	}else{
		
		if (len(vmID) > 0){
			if (latestPerVM == "true"){
				db.Where("vm_id = ? and created_at > ? and created_at < ?", vmID, toDate).Order("created_at desc").First(&reportEntities)
				return getReportModels(reportEntities)
			}else{
				db.Where("vm_id = ? and created_at > ? and created_at < ?", vmID, toDate).Order("created_at desc").Find(&reportEntities)
				return getReportModels(reportEntities)
			}
			
		}
				
		if (len(hardwareUUID) > 0){
			if (latestPerVM == "true"){
				db.Table("report_entities").Joins("LEFT JOIN report_entities as r2 ON report_entities.vm_id = r2.vm_id AND report_entities.created_at > r2.created_at").Where("r2.vm_id is null AND report_entities.trust_report->'vm_manifest'->'vm_info'->>'host_hardware_uuid' = ? AND report_entities.created_at < ?", hardwareUUID, toDate).Find(&reportEntities)
				return getReportModels(reportEntities)
			}else{
				db.Order("created_at desc").Where("trust_report->'vm_manifest'->'vm_info'->>'host_hardware_uuid' = ? AND created_at < ?", hardwareUUID, toDate).Find(&reportEntities)
				return getReportModels(reportEntities)
			}
		}

		// If only to_date is the given filter criteria
		if (latestPerVM == "true"){
			db.Joins("LEFT JOIN report_entities as r2 ON report_entities.vm_id = r2.vm_id AND report_entities.created_at > r2.created_at").Where("r2.vm_id is null AND report_entities.created_at < ?", toDate).Find(&reportEntities)
			fmt.Println("only to date")
			return getReportModels(reportEntities)
		}else{
			db.Order("created_at desc").Where("created_at < ?", toDate).Find(&reportEntities)
			return getReportModels(reportEntities)
		}

	}
	return getReportModels(reportEntities)
}

func (repo *reportRepo) Create(vtr *model.VMTrustReport) error {
	if vtr == nil {
		return errors.New("cannot create nil report")
	}
	id, err := uuid.NewV4()
	if err != nil {
		return errors.New("Unable to create uuid. ")
	}

	//Create report ID for every post request
	var report model.Report
	report.ID = id.String()
	
	tx := repo.db.Begin()
			
	if len(vtr.Manifest.VmInfo.VmID) == 0 && len(vtr.Manifest.VmInfo.HostHardwareUUID) == 0 && len(vtr.Manifest.VmInfo.ImageID) == 0{
		tx.Rollback()
		return errors.New("Instance uuid cannot be empty")
	}

	if err := tx.Create(&reportEntity{VMTrustReport: vtr, Report: report}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (repo *reportRepo) DeleteByReportID(uuid string) error {
	// Delete 
	return repo.db.Where("id = ?", uuid).Delete(reportEntity{}).Error
}

// GetReportRepository gets a Repository connector for the supplied gorm DB instance
func GetReportRepository(db *gorm.DB) ReportRepository {
	db.AutoMigrate(&reportEntity{})
	repo := &reportRepo{
		db: db,
	}
	return repo
}
