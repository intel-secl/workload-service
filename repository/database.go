package repository

import (
	"github.com/jinzhu/gorm"
)

type WlsDatabase interface {
	Migrate() error
	FlavorRepository() FlavorRepository
	ImageRepository() ImageRepository
	Driver() *gorm.DB
}
