/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package postgres

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
	"intel/isecl/workload-service/v4/model"
	"intel/isecl/workload-service/v4/repository"
	"strconv"
	"time"
)

type reportRepo struct {
	db *gorm.DB
}

const dateString string = "2006-01-02T15:04:05"
const dateFormatString string = "2006-01-02 15:04:05"

func parseTime(strTime string) (time.Time, error) {
	log.Trace("repository/postgres/report_repository:parseTime() Entering")
	defer log.Trace("repository/postgres/report_repository:parseTime() Leaving")

	t, err := time.Parse(dateString, strTime)
	if err != nil {
		return t, errors.Wrap(err, "repository/postgres/report_repository:parseTime() Invalid date format, should be yyyy-mm-ddThh:mm:ss")
	}
	return t, nil
}

func getReportModels(reportEntities []reportEntity) ([]model.Report, error) {
	log.Trace("repository/postgres/report_repository:getReportModels() Entering")
	defer log.Trace("repository/postgres/report_repository:getReportModels() Leaving")

	ids := make([]model.Report, len(reportEntities))
	for i, v := range reportEntities {
		ids[i] = v.Report()
	}
	return ids, nil
}

func (repo reportRepo) RetrieveByFilterCriteria(filter repository.ReportFilter) ([]model.Report, error) {
	log.Trace("repository/postgres/report_repository:RetrieveByFilterCriteria() Entering")
	defer log.Trace("repository/postgres/report_repository:RetrieveByFilterCriteria() Leaving")

	db := repo.db
	var reportEntities []reportEntity
	var err error

	instanceID := ""
	reportID := ""
	hardwareUUID := ""
	var toDate time.Time
	var fromDate time.Time
	latestPerVM := true
	filterQuery := true

	if len(filter.InstanceID) > 0 {
		instanceID = filter.InstanceID
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
			return nil, errors.Wrap(err, "Invalid date format, should be yyyy-mm-ddThh:mm:ss")
		}
	}

	if len(filter.FromDate) > 0 {
		fromDate, err = parseTime(filter.FromDate)
		if err != nil {
			return nil, errors.Wrap(err, "Invalid date format, should be yyyy-mm-ddThh:mm:ss")
		}
	}

	if filter.NumOfDays > 0 {
		toDate, err = parseTime(time.Now().Format(dateString))
		if err != nil {
			return nil, errors.Wrap(err, "Invalid date format, should be yyyy-mm-ddThh:mm:ss")
		}

		fromDate, err = parseTime(toDate.AddDate(0, 0, -(filter.NumOfDays)).Format(dateString))
		if err != nil {
			return nil, errors.Wrap(err, "Invalid date format, should be yyyy-mm-ddThh:mm:ss")
		}
	}

	if len(filter.LatestPerVM) > 0 {
		latestPerVM, _ = strconv.ParseBool(filter.LatestPerVM)
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
	return findReports(instanceID, hardwareUUID, toDate, fromDate, latestPerVM, db)
}

func findReports(instanceID string, hardwareUUID string, toDate time.Time, fromDate time.Time, latestPerVM bool, db *gorm.DB) ([]model.Report, error) {
	log.Trace("repository/postgres/report_repository:findReports() Entering")
	defer log.Trace("repository/postgres/report_repository:findReports() Leaving")

	var reportEntities []reportEntity
	partialQueryString := ""

	if instanceID != "" {
		partialQueryString = fmt.Sprintf("trust_report -> 'instance_manifest' -> 'instance_info' ->> 'instance_id' = '%s'", instanceID)
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
	log.Trace("repository/postgres/report_repository:Create() Entering")
	defer log.Trace("repository/postgres/report_repository:Create() Leaving")

	if report == nil {
		return errors.New("repository/postgres/report_repository:Create() cannot create nil report")
	}
	if len(report.Manifest.InstanceInfo.InstanceID) == 0 && len(report.Manifest.InstanceInfo.HostHardwareUUID) == 0 && len(report.Manifest.InstanceInfo.ImageID) == 0 {
		return errors.New("repository/postgres/report_repository:Create() instance uuid cannot be empty")
	}
	reportJSON, err := json.Marshal(report.InstanceTrustReport)
	if err != nil {
		return errors.Wrap(err, "repository/postgres/report_repository:Create() failed to marshal instance trust report to JSON")
	}
	signedJSON, err := json.Marshal(report.SignedData)
	if err != nil {
		return errors.Wrap(err, "repository/postgres/report_repository:Create() failed to marshal signed data to JSON")
	}
	if err := repo.db.Create(
		&reportEntity{
			TrustReport: postgres.Jsonb{RawMessage: reportJSON},
			SignedData:  postgres.Jsonb{RawMessage: signedJSON},
			InstanceID:  report.Manifest.InstanceInfo.InstanceID,
		}).Error; err != nil {
		return errors.Wrap(err, "repository/postgres/report_repository:Create() Failed to create instance trust report")
	}
	return nil
}

func (repo reportRepo) DeleteByReportID(uuid string) error {
	log.Trace("repository/postgres/report_repository:DeleteByReportID() Entering")
	defer log.Trace("repository/postgres/report_repository:DeleteByReportID() Leaving")
	return repo.db.Delete(&reportEntity{ID: uuid}).Error
}
