package postgres

import (
	"errors"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"

	"github.com/jinzhu/gorm"
)

type imageRepo struct {
	db *gorm.DB
}

func (repo imageRepo) RetrieveByFilterCriteria(locator repository.ImageLocator) ([]model.Image, error) {
	db := repo.db
	if len(locator.FlavorID) > 0 {
		db = db.Joins("LEFT JOIN flavors ON flavors.id = ?", locator.FlavorID)
	}

	var entities []imageEntity
	err := db.Preload("Flavors").Find(&entities).Error

	if err != nil {
		return nil, err
	}

	images := make([]model.Image, len(entities))
	for i, v := range entities {
		images[i] = v.Image()
	}
	return images, nil
}

// RetrieveAssociatedFlavors returns a list of FlavorIDs associated with the specified image uuid
func (repo imageRepo) RetrieveAssociatedFlavors(uuid string) ([]string, error) {
	i := imageEntity{}
	if err := repo.db.Preload("Flavors").First(&i, "id = ?", uuid).Error; err != nil {
		return nil, err
	}
	return i.Image().FlavorIDs, nil
}

func (repo imageRepo) Create(image *model.Image) error {
	tx := repo.db.Begin()
	ie := imageEntity{ID: image.ID}
	if !tx.Take(&ie).RecordNotFound() {
		// already exists
		tx.Rollback()
		return repository.ErrImageAssociationAlreadyExists
	}
	// make sure the list of flavorID's makes sense, this *could* be done at the Database schema level as an optimization
	var count int
	tx.Model(&flavorEntity{}).Where("id in (?)", image.FlavorIDs).Count(&count)
	if count != len(image.FlavorIDs) {
		// some flavor ID's dont exist
		return errors.New("one or more FlavorID's does not exist in the database")
	}
	flavorEntities := make([]flavorEntity, len(image.FlavorIDs))
	for i, id := range image.FlavorIDs {
		flavorEntities[i] = flavorEntity{ID: id}
	}
	ie.Flavors = flavorEntities
	err := tx.Create(&ie).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (repo imageRepo) RetrieveByUUID(uuid string) (*model.Image, error) {
	var i imageEntity
	res := repo.db.Preload("Flavors").First(&i, "id = ?", uuid)
	if res.Error != nil {
		return nil, res.Error
	}
	image := i.Image()
	return &image, nil
}

func (repo imageRepo) DeleteByUUID(uuid string) error {
	return repo.db.Delete(imageEntity{}, "id = ?", uuid).Error
}
