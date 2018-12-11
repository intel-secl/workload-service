package repository

import (
	"errors"

	"github.com/jinzhu/gorm"
)

var (
	ErrImageAssociationAlreadyExists = errors.New("image association with UUID already exists")
)

// ImageLocator specifies query filter criteria for retrieving images. Each Field may be empty
type ImageLocator struct {
	ImageID  string `json:"image_id, omitempty"`
	FlavorID string `json:"flavor_id, omitempty"`
}

// ImageFlavorRepository defines an interface that provides persistence operations for an Image-Flavor link.
// It defines High Level CRUD operations that could be implemented by any database or persistence layer (such as postgres)
// The CRUD operations are logically grouped, but not defined to any single interface, so that FlavorRepository may customize them to its own needs, with
// Stronger typing rather than cast everything from an interface{}
type ImageFlavorRepository interface {
	// C
	Create(imageUUID string, flavorUUID string) error
	// R
	RetrieveByUUID(uuid string) (bool, error)
	RetrieveByFilterCriteria(locator ImageLocator) ([]string, error)
	// D
	DeleteByUUID(uuid string) error
}

type imageFlavorRepo struct {
	db *gorm.DB
}

func (ifr *imageFlavorRepo) RetrieveByFilterCriteria(locator ImageLocator) ([]string, error) {
	db := ifr.db
	if len(locator.ImageID) > 0 {
		db = db.Where("image_id = ?", locator.ImageID)
	}
	if len(locator.FlavorID) > 0 {
		db = db.Where("flavor_id = ?", locator.FlavorID)
	}
	var entities []imageFlavorEntity
	err := db.Find(&entities).Error

	if err != nil {
		return nil, err
	}

	ids := make([]string, len(entities))
	for i, v := range entities {
		ids[i] = v.ID
	}
	return ids, nil
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
