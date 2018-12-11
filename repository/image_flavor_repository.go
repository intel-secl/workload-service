package repository

import (
	"errors"

	"github.com/jinzhu/gorm"
)

var (
	ErrImageAssociationAlreadyExists = errors.New("image association with UUID already exists")
)

// ImageFlavorRepository defines an interface that provides persistence operations for an Image-Flavor link.
// It defines High Level CRUD operations that could be implemented by any database or persistence layer (such as postgres)
// The CRUD operations are logically grouped, but not defined to any single interface, so that FlavorRepository may customize them to its own needs, with
// Stronger typing rather than cast everything from an interface{}
type ImageFlavorRepository interface {
	// C
	Create(imageUUID string, flavorUUID string) error
	// R
	RetrieveByUUID(uuid string) (bool, error)
	// D
	DeleteByUUID(uuid string) error
}

type imageFlavorRepo struct {
	db *gorm.DB
}

func (ifr *imageFlavorRepo) Create(imageUUID string, flavorUUID string) error {
	tx := ifr.db.Begin()
	var i imageFlavorEntity
	if !tx.Take(&i, "id = ?", imageUUID).RecordNotFound() {
		// already exists
		tx.Rollback()
		return ErrImageAssociationAlreadyExists
	} else {
		err := tx.Create(&imageFlavorEntity{ID: imageUUID, FlavorID: flavorUUID}).Error
		if err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
		return nil
	}
}

func (ifr *imageFlavorRepo) RetrieveByUUID(uuid string) (bool, error) {
	var i imageFlavorEntity
	res := ifr.db.First(&i, "id = ?", uuid)
	if res.Error != nil {
		if res.RecordNotFound() {
			return false, nil
		} else {
			return false, res.Error
		}
	} else {
		return true, nil
	}
}

func (ifr *imageFlavorRepo) DeleteByUUID(uuid string) error {
	return ifr.db.Delete(&imageFlavorEntity{ID: uuid}).Error
}

// GetImageFlavorRepository gets a Repository connector for the supplied gorm DB instance
func GetImageFlavorRepository(db *gorm.DB) ImageFlavorRepository {
	db.AutoMigrate(&imageFlavorEntity{})
	repo := &imageFlavorRepo{
		db: db,
	}
	return repo
}
