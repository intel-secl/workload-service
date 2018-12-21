package repository

import (
	"github.com/jinzhu/gorm"
)

type WlsDatabase interface {
	Migrate() error
	FlavorRepository() FlavorRepository
	ImageRepository() ImageRepository
	ReportRepository() ReportRepository
	Driver() *gorm.DB
}
