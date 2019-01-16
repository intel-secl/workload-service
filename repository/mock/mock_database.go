package mock

import (
	"intel/isecl/workload-service/repository"

	"github.com/jinzhu/gorm"
)

// Database provides a mock Db
type Database struct {
	MockFlavor MockFlavor
	MockImage  MockImage
	MockReport MockReport
}

func (m *Database) Migrate() error {
	return nil
}

func (m *Database) FlavorRepository() repository.FlavorRepository {
	return &m.MockFlavor
}

func (m *Database) ImageRepository() repository.ImageRepository {
	return &m.MockImage
}

func (m *Database) ReportRepository() repository.ReportRepository {
	return &m.MockReport
}

func (m *Database) Driver() *gorm.DB {
	return nil
}
