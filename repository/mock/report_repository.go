/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package mock

import (
	"intel/isecl/workload-service/v3/model"
	"intel/isecl/workload-service/v3/repository"
)

type MockReport struct {
	CreateFn                   func(*model.Report) error
	RetrieveByFilterCriteriaFn func(repository.ReportFilter) ([]model.Report, error)
	DeleteByReportIDFn         func(string) error
}

func (m *MockReport) Create(r *model.Report) error {
	log.Trace("repository/mock/report_repository:Create() Entering")
	defer log.Trace("repository/mock/report_repository:Create() Leaving")
	log.Debug("repository/mock/report_repository:Create() Create mock report")
	if m.CreateFn != nil {
		return m.CreateFn(r)
	}
	return nil
}

func (m *MockReport) RetrieveByFilterCriteria(filter repository.ReportFilter) ([]model.Report, error) {
	log.Trace("repository/mock/report_repository:RetrieveByFilterCriteria() Entering")
	defer log.Trace("repository/mock/report_repository:RetrieveByFilterCriteria() Leaving")
	log.Debug("repository/mock/report_repository:RetrieveByFilterCriteria() Retrieve mock report by filter criteria")
	if m.RetrieveByFilterCriteriaFn != nil {
		return m.RetrieveByFilterCriteriaFn(filter)
	}
	return []model.Report{r}, nil
}

func (m *MockReport) DeleteByReportID(reportID string) error {
	log.Trace("repository/mock/report_repository:DeleteByReportID() Entering")
	defer log.Trace("repository/mock/report_repository:DeleteByReportID() Leaving")
	log.Debug("repository/mock/report_repository:DeleteByReportID() Delete mock report by Report ID")
	if m.DeleteByReportIDFn != nil {
		return m.DeleteByReportIDFn(reportID)
	}
	return nil
}
