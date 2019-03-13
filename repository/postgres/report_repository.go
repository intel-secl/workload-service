package postgres

import (
	"encoding/json"
	"errors"
	"fmt"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type reportRepo struct {
	db *gorm.DB
}

const dateString string = "2006-01-02 15:04:05.999999-07"

func parseTime(strTime string) time.Time {
	t, err := time.Parse(dateString, strTime)
	if err != nil {
		fmt.Println(err)
	}
	return t
}

func getReportModels(reportEntities []reportEntity) ([]model.Report, error) {
	ids := make([]model.Report, len(reportEntities))

	for i, v := range reportEntities {
		ids[i] = v.Report()
	}
	fmt.Println(len(ids))
	return ids, nil
}

func (repo reportRepo) RetrieveByFilterCriteria(filter repository.ReportFilter) ([]model.Report, error) {
	db := repo.db
	var reportEntities []reportEntity

	vmID := ""
	reportID := ""
	hardwareUUID := ""
	var toDate time.Time
	var fromDate time.Time
	latestPerVM := "true"

	if len(filter.VMID) > 0 {
		vmID = filter.VMID
	}

	if len(filter.ReportID) > 0 {
		reportID = filter.ReportID
	}

	if len(filter.HardwareUUID) > 0 {
		hardwareUUID = filter.HardwareUUID
	}

	if len(filter.ToDate) > 0 {
		toDate = parseTime(filter.ToDate)
	}

	if len(filter.FromDate) > 0 {
		fromDate = parseTime(filter.FromDate)
	}

	if filter.NumOfDays > 0 {
		toDate = parseTime(time.Now().Format(dateString))
		fromDate = parseTime(toDate.AddDate(0, 0, -(filter.NumOfDays)).Format(dateString))
	}

	if len(filter.LatestPerVM) > 0 {
		latestPerVM = filter.LatestPerVM
	}

	//Only fetch the report since reportid is unique across the table
	if len(reportID) > 0 {
		db.Where("id = ?", reportID).Find(&reportEntities)
		return getReportModels(reportEntities)
	}
	fmt.Println("todate")
	fmt.Println(toDate)
	fmt.Println("todate")
	if latestPerVM == "true" && toDate.IsZero() && fromDate.IsZero() {
		return findLatestReports(vmID, reportID, hardwareUUID, db)
	}
	return findReports(vmID, reportID, hardwareUUID, toDate, fromDate, latestPerVM, db)
}

func findLatestReports(vmID string, reportID string, hardwareUUID string, db *gorm.DB) ([]model.Report, error) {

	var reportEntities []reportEntity
	//Only fetch latest record for given a vm_id
	if len(vmID) > 0 {
		db.Where("vm_id = ?", vmID).Order("created_at desc").First(&reportEntities)
		return getReportModels(reportEntities)
	}

	if len(hardwareUUID) > 0 {
		// what is this JOIN for? it seems at a glance that it's just an identity operation
		db.Joins("LEFT JOIN reports as r2 ON reports.vm_id = r2.vm_id AND reports.created_at > r2.created_at").Where("r2.vm_id is null and reports.trust_report->'vm_manifest'->'vm_info'->>'host_hardware_uuid' = ?", hardwareUUID).Find(&reportEntities)
		return getReportModels(reportEntities)
	}
	return getReportModels(reportEntities)
}

func findReports(vmID string, reportID string, hardwareUUID string, toDate time.Time, fromDate time.Time, latestPerVM string, db *gorm.DB) ([]model.Report, error) {
	var reportEntities []reportEntity

	if !fromDate.IsZero() && !toDate.IsZero() {

		if len(vmID) > 0 {
			if latestPerVM == "true" {
				db.Where("vm_id = ? AND created_at > ? AND created_at < ?", vmID, fromDate, toDate).Order("created_at desc").First(&reportEntities)
				return getReportModels(reportEntities)
			}
			db.Where("vm_id = ? AND created_at > ? AND created_at < ?", vmID, fromDate, toDate).Order("created_at desc").Find(&reportEntities)
			return getReportModels(reportEntities)
		}

		if len(hardwareUUID) > 0 {
			if latestPerVM == "true" {
				db.Joins("LEFT JOIN reports as r2 ON reports.vm_id = r2.vm_id AND reports.created_at > r2.created_at").Where("r2.vm_id is null AND reports.trust_report->'vm_manifest'->'vm_info'->>'host_hardware_uuid' = ? AND reports.created_at > ? and reports.created_at < ?", hardwareUUID, fromDate, toDate).Find(&reportEntities)
				return getReportModels(reportEntities)
			}
			db.Order("created_at desc").Where("trust_report->'vm_manifest'->'vm_info'->>'host_hardware_uuid' = ? AND created_at > ? and created_at < ?", hardwareUUID, fromDate, toDate).Find(&reportEntities)
			return getReportModels(reportEntities)
		}
		// If only either num of days or from_date and to_date is the given filter criteria
		if latestPerVM == "true" {
			db.Joins("LEFT JOIN reports as r2 ON reports.vm_id = r2.vm_id AND reports.created_at > r2.created_at").Where("r2.vm_id is null AND reports.created_at > ? and reports.created_at < ?", fromDate, toDate).Find(&reportEntities)
			fmt.Println("only num of days")
			return getReportModels(reportEntities)
		}
		db.Order("created_at desc").Where("created_at > ? and created_at < ?", fromDate, toDate).Find(&reportEntities)
		return getReportModels(reportEntities)
	}

	if !fromDate.IsZero() {
		if len(vmID) > 0 {
			if latestPerVM == "true" {
				db.Where("vm_id = ? and created_at > ? ", vmID, fromDate).Order("created_at desc").First(&reportEntities)
				return getReportModels(reportEntities)
			}
			db.Where("vm_id = ? and created_at > ? ", vmID, fromDate).Order("created_at desc").Find(&reportEntities)
			return getReportModels(reportEntities)
		}

		if len(hardwareUUID) > 0 {
			fmt.Println("I am heere")
			if latestPerVM == "true" {
				db.Table("reports").Joins("LEFT JOIN reports as r2 ON reports.vm_id = r2.vm_id AND reports.created_at > r2.created_at").Where("r2.vm_id is null AND reports.trust_report->'vm_manifest'->'vm_info'->>'host_hardware_uuid' = ? AND created_at > ?", hardwareUUID, fromDate).Find(&reportEntities)
				return getReportModels(reportEntities)
			}
			db.Order("created_at desc").Where("trust_report->'vm_manifest'->'vm_info'->>'host_hardware_uuid' = ? and created_at > ? ", hardwareUUID, fromDate).Find(&reportEntities)
			return getReportModels(reportEntities)
		}

		// If only from_date is the given filter criteria
		if latestPerVM == "true" {
			fmt.Println("from date")
			db.Joins("LEFT JOIN reports as r2 ON reports.vm_id = r2.vm_id AND reports.created_at > r2.created_at").Where("r2.vm_id is null AND reports.created_at > ?", fromDate).Find(&reportEntities)
			return getReportModels(reportEntities)
		}
		db.Order("created_at desc").Where("created_at > ?", fromDate).Find(&reportEntities)
		return getReportModels(reportEntities)
	}
	if len(vmID) > 0 {
		if latestPerVM == "true" {
			db.Where("vm_id = ? and created_at > ? and created_at < ?", vmID, toDate).Order("created_at desc").First(&reportEntities)
			return getReportModels(reportEntities)
		}
		db.Where("vm_id = ? and created_at > ? and created_at < ?", vmID, toDate).Order("created_at desc").Find(&reportEntities)
		return getReportModels(reportEntities)
	}

	if len(hardwareUUID) > 0 {
		if latestPerVM == "true" {
			db.Table("reports").Joins("LEFT JOIN reports as r2 ON reports.vm_id = r2.vm_id AND reports.created_at > r2.created_at").Where("r2.vm_id is null AND reports.trust_report->'vm_manifest'->'vm_info'->>'host_hardware_uuid' = ? AND reports.created_at < ?", hardwareUUID, toDate).Find(&reportEntities)
			return getReportModels(reportEntities)
		}
		db.Order("created_at desc").Where("trust_report->'vm_manifest'->'vm_info'->>'host_hardware_uuid' = ? AND created_at < ?", hardwareUUID, toDate).Find(&reportEntities)
		return getReportModels(reportEntities)
	}

	// If only to_date is the given filter criteria
	if latestPerVM == "true" {
		db.Joins("LEFT JOIN reports as r2 ON reports.vm_id = r2.vm_id AND reports.created_at > r2.created_at").Where("r2.vm_id is null AND reports.created_at < ?", toDate).Find(&reportEntities)
		fmt.Println("only to date")
		return getReportModels(reportEntities)
	}
	db.Order("created_at desc").Where("created_at < ?", toDate).Find(&reportEntities)
	return getReportModels(reportEntities)
}

func (repo reportRepo) Create(report *model.Report) error {
	if report == nil {
		return errors.New("cannot create nil report")
	}
	if len(report.Manifest.InstanceInfo.InstanceID) == 0 && len(report.Manifest.InstanceInfo.HostHardwareUUID) == 0 && len(report.Manifest.InstanceInfo.ImageID) == 0 {
		return errors.New("Instance uuid cannot be empty")
	}
	reportJSON, err := json.Marshal(report.InstanceTrustReport)
	if err != nil {
		return err
	}
	signedJSON, err := json.Marshal(report.SignedData)
	if err != nil {
		return err
	}
	if err := repo.db.Create(
		&reportEntity{
			TrustReport: postgres.Jsonb{RawMessage: reportJSON},
			SignedData:  postgres.Jsonb{RawMessage: signedJSON},
			VMID:        report.Manifest.InstanceInfo.InstanceID,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (repo reportRepo) DeleteByReportID(uuid string) error {
	return repo.db.Delete(&reportEntity{ID: uuid}).Error
}
