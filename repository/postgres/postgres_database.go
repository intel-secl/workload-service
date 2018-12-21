package postgres

import (
	"intel/isecl/workload-service/repository"

	"github.com/jinzhu/gorm"
)

type PostgresDatabase struct {
	DB *gorm.DB
}

func (pd PostgresDatabase) Migrate() error {
	pd.DB.AutoMigrate(&flavorEntity{}, &imageEntity{}, &reportEntity{})
	pd.DB.Table("image_flavors").
		AddForeignKey("image_id", "images(id)", "CASCADE", "CASCADE").
		AddForeignKey("flavor_id", "flavors(id)", "CASCADE", "CASCADE").
		AddUniqueIndex("image_flavor_index", "image_id", "flavor_id")
	return nil
}

func (pd PostgresDatabase) Driver() *gorm.DB {
	return pd.DB
}

func (pd PostgresDatabase) FlavorRepository() repository.FlavorRepository {
	return flavorRepo{db: pd.DB}
}

func (pd PostgresDatabase) ReportRepository() repository.ReportRepository {
	return reportRepo{db: pd.DB}
}

func (pd PostgresDatabase) ImageRepository() repository.ImageRepository {
	return imageRepo{db: pd.DB}
}
