package mock

import (
	"intel/isecl/workload-service/repository"

	"github.com/jinzhu/gorm"
)

// DatabaseMock provides a mock Db
type DatabaseMock struct{}

func (m DatabaseMock) Migrate() error {
	return nil
}

func (m DatabaseMock) FlavorRepository() repository.FlavorRepository {
	return new(mockFlavor)
}

func (m DatabaseMock) ImageRepository() repository.ImageRepository {
	return new(mockImage)
}

func (m DatabaseMock) ReportRepository() repository.ReportRepository {
	return new(mockReport)
}

func (m DatabaseMock) Driver() *gorm.DB {
	return nil
}
