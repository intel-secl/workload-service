/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
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

const dateString string = "2006-01-02T15:04:05"
const dateFormatString string = "2006-01-02 15:04:05"

func parseTime(strTime string) (time.Time, error) {
	t, err := time.Parse(dateString, strTime)
	if err != nil {
		return t, errors.New("Invalid date format, should be yyyy-mm-ddThh:mm:ss")
	}
	return t, nil
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
	var err error

	vmID := ""
	reportID := ""
	hardwareUUID := ""
	var toDate time.Time
	var fromDate time.Time
	latestPerVM := true
	filterQuery := true

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
		toDate, err = parseTime(filter.ToDate)
		if err != nil {
			return nil, err
		}
	}

	if len(filter.FromDate) > 0 {
		fromDate, err = parseTime(filter.FromDate)
		if err != nil {
			return nil, err
		}
	}

	if filter.NumOfDays > 0 {
		toDate, err = parseTime(time.Now().Format(dateString))
		if err != nil {
			return nil, err
		}

		fromDate, err = parseTime(toDate.AddDate(0, 0, -(filter.NumOfDays)).Format(dateString))
		if err != nil {
			return nil, err
		}
	}

	if !filter.LatestPerVM {
		latestPerVM = filter.LatestPerVM
	}

	if !filter.Filter {
		filterQuery = filter.Filter
	}

	//Only fetch the report since reportid is unique across the table
	if len(reportID) > 0 {
		db.Where("id = ?", reportID).Find(&reportEntities)
		return getReportModels(reportEntities)
	}

	// fetch all the reports if filter=false
	if !filterQuery {
		db.Find(&reportEntities)
		return getReportModels(reportEntities)
	}

	return findReports(vmID, hardwareUUID, toDate, fromDate, latestPerVM, db)
}

func findReports(vmID string, hardwareUUID string, toDate time.Time, fromDate time.Time, latestPerVM bool, db *gorm.DB) ([]model.Report, error) {
	var reportEntities []reportEntity
	partialQueryString := ""

	if vmID != "" {
		partialQueryString = fmt.Sprintf("vm_id = '%s'", vmID)
	}

	if hardwareUUID != "" {
        if partialQueryString != "" {
			partialQueryString = fmt.Sprintf("%s AND trust_report -> 'instance_manifest' -> 'instance_info' ->> 'host_hardware_uuid' = '%s'", partialQueryString, hardwareUUID)
		} else {
			partialQueryString = fmt.Sprintf("trust_report -> 'instance_manifest' -> 'instance_info' ->> 'host_hardware_uuid' = '%s'", hardwareUUID)
		}
	}

	if !fromDate.IsZero() {
		if partialQueryString != "" {
			partialQueryString = fmt.Sprintf("%s AND created_at >= CAST('%s' AS TIMESTAMP)", partialQueryString, fromDate.Format(dateFormatString))
		} else {
			partialQueryString = fmt.Sprintf("created_at >= CAST('%s' AS TIMESTAMP)", fromDate.Format(dateFormatString))
		}
	}

	if !toDate.IsZero() {
		if partialQueryString != "" {
			partialQueryString = fmt.Sprintf("%s AND created_at <= CAST('%s' AS TIMESTAMP)", partialQueryString, toDate.Format(dateFormatString))
		} else {
			partialQueryString = fmt.Sprintf("created_at <= CAST('%s' AS TIMESTAMP)", toDate.Format(dateFormatString))
		}
	}

	if latestPerVM {
		db.Where(partialQueryString).Order("created_at desc").First(&reportEntities)
		return getReportModels(reportEntities)
	}
	
	db.Where(partialQueryString).Find(&reportEntities)
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
