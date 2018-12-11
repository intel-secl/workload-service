package repository

import (
	"errors"
	"intel/isecl/workload-service/model"

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

// ImageRepository defines an interface that provides persistence operations for an Image-Flavor link.
// It defines High Level CRUD operations that could be implemented by any database or persistence layer (such as postgres)
// The CRUD operations are logically grouped, but not defined to any single interface, so that FlavorRepository may customize them to its own needs, with
// Stronger typing rather than cast everything from an interface{}
type ImageRepository interface {
	// C
	Create(image *model.Image) error
	// R
	RetrieveByUUID(uuid string) (*model.Image, error)
	RetrieveByFilterCriteria(locator ImageLocator) ([]model.Image, error)
	// D
	DeleteByUUID(uuid string) error
}

type imageRepo struct {
	db *gorm.DB
}

func (ifr *imageRepo) RetrieveByFilterCriteria(locator ImageLocator) ([]model.Image, error) {
	db := ifr.db
	if len(locator.ImageID) > 0 {
		db = db.Where("image_id = ?", locator.ImageID)
	}
	if len(locator.FlavorID) > 0 {
		db = db.Where("flavor_id = ?", locator.FlavorID)
	}
	var entities []imageEntity
	err := db.Find(&entities).Error

	if err != nil {
		return nil, err
	}

	ids := make([]model.Image, len(entities))
	for i, v := range entities {
		ids[i] = model.Image{ID: v.ID, FlavorID: v.FlavorID}
	}
	return ids, nil
}

func (ifr *imageRepo) Create(image *model.Image) error {
	tx := ifr.db.Begin()
	var i imageEntity
	if !tx.Take(&i, "id = ?", image.ID).RecordNotFound() {
		// already exists
		tx.Rollback()
		return ErrImageAssociationAlreadyExists
	}
	err := tx.Create(&imageEntity{ID: image.ID, FlavorID: image.FlavorID}).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (ifr *imageRepo) RetrieveByUUID(uuid string) (*model.Image, error) {
	var i imageEntity
	res := ifr.db.First(&i, "id = ?", uuid)
	if res.Error != nil {
		return nil, res.Error
	} else {
		return &model.Image{ID: i.ID, FlavorID: i.FlavorID}, nil
	}
}

func (ifr *imageRepo) DeleteByUUID(uuid string) error {
	return ifr.db.Delete(&imageEntity{ID: uuid}).Error
}

// GetImageFlavorRepository gets a Repository connector for the supplied gorm DB instance
func GetImageFlavorRepository(db *gorm.DB) ImageRepository {
	db.AutoMigrate(&imageEntity{})
	repo := &imageRepo{
		db: db,
	}
	return repo
}
