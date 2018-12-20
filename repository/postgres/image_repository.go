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

func (repo imageRepo) RetrieveByFilterCriteria(filter repository.ImageFilter) ([]model.Image, error) {
	db := repo.db
	if len(filter.FlavorID) > 0 {
		// find all images that at least contain FlavorID as one of the associations
		db = db.Joins("LEFT JOIN image_flavors ON (image_flavors.image_id = images.id AND image_flavors.flavor_id = ?)", filter.FlavorID)
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

func (repo imageRepo) RetrieveAssociatedFlavor(imageUUID string, flavorUUID string) (*model.Flavor, error) {
	var image imageEntity
	if err := repo.db.Preload("Flavors", "id = ?", flavorUUID).First(&image, "id = ?", imageUUID).Error; err != nil {
		return nil, err
	}
	return nil, nil
}

func (repo imageRepo) RetrieveAssociatedFlavors(uuid string) ([]model.Flavor, error) {
	var image imageEntity
	if err := repo.db.Preload("Flavors").First(&image, "id = ?", uuid).Error; err != nil {
		return nil, err
	}
	flavors := make([]model.Flavor, len(image.Flavors))
	for i, f := range image.Flavors {
		flavors[i] = f.Flavor()
	}
	return flavors, nil
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
