package postgres

import (
	"intel/isecl/workload-service/repository"

	"github.com/jinzhu/gorm"
)

type PostgresDatabase struct {
	DB *gorm.DB
}

func (pd PostgresDatabase) Migrate() error {
	pd.DB.AutoMigrate(&flavorEntity{}, &imageEntity{})
	pd.DB.Table("image_flavors").AddForeignKey("image_id", "images(id)", "CASCADE", "CASCADE").AddForeignKey("flavor_id", "flavors(id)", "CASCADE", "CASCADE")
	return nil
}

func (pd PostgresDatabase) FlavorRepository() repository.FlavorRepository {
	return flavorRepo{db: pd.DB}
}

func (pd PostgresDatabase) ImageRepository() repository.ImageRepository {
	return imageRepo{db: pd.DB}
}
