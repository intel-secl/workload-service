package repository

import (
	"errors"
	"intel/isecl/lib/flavor"

	"github.com/jinzhu/gorm"
)

var (
	ErrFlavorUUIDAlreadyExists  = errors.New("flavor already exists with UUID")
	ErrFlavorLabelAlreadyExists = errors.New("flavor already exists with label")
)

// FlavorRepository defines an interface that provides persistence operations for a Flavor.
// It defines High Level CRUD operations that could be implemented by any database or persistence layer (such as postgres)
// The CRUD operations are logically grouped, but not defined to any single interface, so that FlavorRepository may customize them to its own needs, with
// Stronger typing rather than cast everything from an interface{}
type FlavorRepository interface {
	// C
	Create(f *flavor.ImageFlavor) error
	// R
	RetrieveByUUID(uuid string) (*flavor.ImageFlavor, error)
	RetrieveByLabel(label string) (*flavor.ImageFlavor, error)
	// D
	Delete(f *flavor.ImageFlavor) error
	DeleteByUUID(uuid string) error
	DeleteByLabel(label string) error
}

type flavorRepo struct {
	db *gorm.DB
}

func (repo *flavorRepo) Create(f *flavor.ImageFlavor) error {
	if f == nil {
		return errors.New("cannot create nil flavor")
	}
	tx := repo.db.Begin()
	var fe flavorEntity
	if !tx.Where("id = ?", f.Image.Meta.ID).Or("label = ?", f.Image.Meta.Description.Label).Take(&fe).RecordNotFound() {
		// duplicate exists
		tx.Rollback()
		if fe.ID == f.Image.Meta.ID {
			return ErrFlavorUUIDAlreadyExists
		} else if fe.Label == f.Image.Meta.Description.Label {
			return ErrFlavorLabelAlreadyExists
		} else {
			// panic since this shouldn't be reached, indicates a critical error in the code/db logic
			panic("This shouldn't be reached, logic error")
		}
	}
	if err := tx.Create(&flavorEntity{ImageFlavor: f}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (repo *flavorRepo) RetrieveByUUID(uuid string) (*flavor.ImageFlavor, error) {
	var fe flavorEntity
	fe.ID = uuid
	if err := repo.db.First(&fe).Error; err != nil {
		return nil, err
	}
	return fe.ImageFlavor, nil
}

func (repo *flavorRepo) RetrieveByLabel(label string) (*flavor.ImageFlavor, error) {
	var fe flavorEntity
	if err := repo.db.Where("label = ?", label).Find(&fe).Error; err != nil {
		return nil, err
	}
	return fe.ImageFlavor, nil
}

func (repo *flavorRepo) Delete(f *flavor.ImageFlavor) error {
	if f == nil {
		return errors.New("cannot delete nil flavor")
	}
	return repo.DeleteByUUID(f.Image.Meta.ID)
}

func (repo *flavorRepo) DeleteByUUID(uuid string) error {
	err := repo.db.Delete(&flavorEntity{ID: uuid}).Error
	if err != nil {
		return err
	}
	// Delete associated images
	return repo.db.Where("flavor_id = ?", uuid).Delete(imageEntity{}).Error
}

func (repo *flavorRepo) DeleteByLabel(label string) error {
	return repo.db.Where("label = ?", label).Delete(&flavorEntity{}).Error
}

// GetFlavorRepository gets a Repository connector for the supplied gorm DB instance
func GetFlavorRepository(db *gorm.DB) FlavorRepository {
	db.AutoMigrate(&flavorEntity{})
	db.AutoMigrate(&imageEntity{})
	repo := &flavorRepo{
		db: db,
	}
	return repo
}
