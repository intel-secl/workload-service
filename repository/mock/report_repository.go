package mock

import (
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
)

type mockReport struct{}

func (m mockReport) Create(*model.Report) error {
	return nil
}

func (m mockReport) RetrieveByFilterCriteria(filter repository.ReportFilter) ([]model.Report, error) {
	return []model.Report{r}, nil
}

func (m mockReport) DeleteByReportID(string) error {
	return nil
}
