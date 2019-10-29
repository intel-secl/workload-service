/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package postgres

import (
	flvr "intel/isecl/lib/flavor"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"

	"github.com/pkg/errors"

	"github.com/jinzhu/gorm"
)

type imageRepo struct {
	db *gorm.DB
}

func (repo imageRepo) RetrieveByFilterCriteria(filter repository.ImageFilter) ([]model.Image, error) {
	log.Trace("repository/postgres/image_repository:RetrieveByFilterCriteria() Entering")
	defer log.Trace("repository/postgres/image_repository:RetrieveByFilterCriteria() Leaving")

	log.Debug("repository/postgres/image_repository:RetrieveByFilterCriteria() Retrieve image by filter criteria")
	db := repo.db
	var entities []imageEntity
	var err error

	// handling for options
	// - neither image_id or flavor_id are provided-> no sql filter to return all records
	// - only image_id is provided --> filter on image_id
	// - only flavor_id is provided --> filter on flavor_id
	// - both image_id and flavor_id are provided --> filter on both
	//
	// When image_id or flavor_id is provided, apply a single 'like' query to account for
	// the last three possibilities (otherwise pass through to 'select *')
	if len(filter.ImageID) > 0 || len(filter.FlavorID) > 0 {

		if len(filter.ImageID) == 0 {
			filter.ImageID = "%"
		}

		if len(filter.FlavorID) == 0 {
			filter.FlavorID = "%"
		}

		db = db.Joins("LEFT JOIN image_flavors ON (image_flavors.image_id = images.id)").Where("flavor_id::text like ? and image_id::text like ?", filter.FlavorID, filter.ImageID)
	}

	err = db.Preload("Flavors").Find(&entities).Error
	if err != nil {
		return nil, errors.Wrap(err, "repository/postgres/image_repository:RetrieveByFilterCriteria() Failed to retrieve image by filter criteria")
	}

	images := make([]model.Image, len(entities))
	for i, v := range entities {
		images[i] = v.Image()
	}
	return images, nil
}

func (repo imageRepo) Create(image *model.Image) error {
	log.Trace("repository/postgres/image_repository:Create() Entering")
	defer log.Trace("repository/postgres/image_repository:Create() Leaving")

	tx := repo.db.Begin()
	ie := imageEntity{ID: image.ID}
	if !tx.Take(&ie).RecordNotFound() {
		// already exists
		tx.Rollback()
		return repository.ErrImageAssociationAlreadyExists
	}

	// make sure there are no duplicates by actually going through the ids
	set := make(map[string]bool)
	for _, id := range image.FlavorIDs {
		if set[id] == true {
			return repository.ErrImageAssociationAlreadyExists
		}
		set[id] = true
	}
	var flavorEntities []flavorEntity
	// make sure the list of flavorID's makes sense, this *could* be done at the Database schema level as an optimization
	tx.Find(&flavorEntities, "id in (?)", image.FlavorIDs)
	if len(flavorEntities) != len(image.FlavorIDs) {
		// some flavor ID's dont exist
		tx.Rollback()
		return repository.ErrImageAssociationFlavorDoesNotExist
	}
	// also make sure there is only ONE flavor with FlavorPart = IMAGE
	var found bool
	for _, fe := range flavorEntities {
		if fe.FlavorPart == "IMAGE" {
			if found == true {
				// we have duplicate IMAGE flavorParts
				tx.Rollback()
				return repository.ErrImageAssociationDuplicateImageFlavor
			}
			found = true
		}
	}
	ie.Flavors = flavorEntities
	err := tx.Create(&ie).Error
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "repository/postgres/image_repository:Create() Failed to create image")
	}
	tx.Commit()
	return nil
}

func (repo imageRepo) RetrieveAssociatedImageFlavor(imageUUID string) (*flvr.SignedImageFlavor, error) {
	log.Trace("repository/postgres/image_repository:RetrieveAssociatedImageFlavor() Entering")
	defer log.Trace("repository/postgres/image_repository:RetrieveAssociatedImageFlavor() Leaving")

	var flavorEntity flavorEntity
	if err := repo.db.Joins("LEFT JOIN image_flavors ON image_flavors.flavor_id = flavors.id").First(&flavorEntity, "image_id = ? AND (flavor_part = ? OR flavor_part = ?)", imageUUID, "IMAGE", "CONTAINER_IMAGE").Error; err != nil {
		return nil, errors.Wrap(err, "repository/postgres/image_repository:RetrieveAssociatedImageFlavor() Failed to retrieve associated image flavor")
	}
	flavor := flavorEntity.Flavor()
	return &flavor, nil
}

func (repo imageRepo) RetrieveAssociatedFlavor(imageUUID string, flavorUUID string) (*model.Flavor, error) {
	log.Trace("repository/postgres/image_repository:RetrieveAssociatedFlavor() Entering")
	defer log.Trace("repository/postgres/image_repository:RetrieveAssociatedFlavor() Leaving")

	var flavorEntity flavorEntity
	var flavor model.Flavor
	if err := repo.db.Joins("LEFT JOIN image_flavors ON image_flavors.flavor_id = flavors.id").First(&flavorEntity, "id = ? AND image_id = ?", flavorUUID, imageUUID).Error; err != nil {
		return nil, errors.Wrap(err, "repository/postgres/image_repository:RetrieveAssociatedFlavor() Failed to retrieve associated image flavor")
	}
	flavor.Image = flavorEntity.Flavor().ImageFlavor
	return &flavor, nil
}

func (repo imageRepo) RetrieveAssociatedFlavorByFlavorPart(imageUUID string, flavorPart string) (*flvr.SignedImageFlavor, error) {
	log.Trace("repository/postgres/image_repository:RetrieveAssociatedFlavorByFlavorPart() Entering")
	defer log.Trace("repository/postgres/image_repository:RetrieveAssociatedFlavorByFlavorPart() Leaving")

	var flavorEntity flavorEntity
	if err := repo.db.Joins("LEFT JOIN image_flavors ON image_flavors.flavor_id = flavors.id").First(&flavorEntity, "image_id = ? AND flavor_part = ?", imageUUID, flavorPart).Error; err != nil {
		return nil, errors.Wrap(err, "repository/postgres/image_repository:RetrieveAssociatedFlavorByFlavorPart() Failed to retrieve associated image flavor by flavor part ")
	}
	flavor := flavorEntity.Flavor()
	return &flavor, nil
}

func (repo imageRepo) RetrieveAssociatedFlavors(uuid string) ([]model.Flavor, error) {
	log.Trace("repository/postgres/image_repository:RetrieveAssociatedFlavors() Entering")
	defer log.Trace("repository/postgres/image_repository:RetrieveAssociatedFlavors() Leaving")

	var image imageEntity
	if err := repo.db.Preload("Flavors").First(&image, "id = ?", uuid).Error; err != nil {
		return make([]model.Flavor, 0), errors.Wrap(err, "repository/postgres/image_repository:RetrieveAssociatedFlavors() Failed to retrieve associated image flavors")
	}
	flavors := make([]model.Flavor, len(image.Flavors))
	for i, f := range image.Flavors {
		flavors[i].Image = f.Flavor().ImageFlavor
	}
	return flavors, nil
}

func (repo imageRepo) RetrieveByUUID(uuid string) (*model.Image, error) {
	log.Trace("repository/postgres/image_repository:RetrieveByUUID() Entering")
	defer log.Trace("repository/postgres/image_repository:RetrieveByUUID() Leaving")

	var i imageEntity
	err := repo.db.Preload("Flavors").First(&i, "id = ?", uuid).Error
	if err != nil {
		return nil, errors.Wrap(err, "repository/postgres/image_repository:RetrieveByUUID() Failed to retrieve image by UUID")
	}
	image := i.Image()
	return &image, nil
}

func (repo imageRepo) Update(image *model.Image) error {
	log.Trace("repository/postgres/image_repository:Update() Entering")
	defer log.Trace("repository/postgres/image_repository:Update() Leaving")

	if image == nil {
		return errors.New("repository/postgres/image_repository:Update() cannot update nil image")
	}
	tx := repo.db.Begin()
	var ie imageEntity
	if err := tx.First(&ie, "id = ?", image.ID).Error; err != nil {
		tx.Rollback()
		return err
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
	err = tx.Save(&ie).Error
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "repository/postgres/image_repository:Update() Failed to update image")
	}
	return tx.Commit().Error
}

func (repo imageRepo) AddAssociatedFlavor(imageUUID string, flavorUUID string) error {
	log.Trace("repository/postgres/image_repository:AddAssociatedFlavor() Entering")
	defer log.Trace("repository/postgres/image_repository:AddAssociatedFlavor() Leaving")

	tx := repo.db.Begin()
	ie := imageEntity{
		ID: imageUUID,
	}
	fe := flavorEntity{
		ID: flavorUUID,
	}
	if err := tx.First(&fe).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Preload("Flavors").First(&ie).Error; err != nil {
		tx.Rollback()
		return err
	}
	assoc := tx.Model(&imageEntity{ID: imageUUID}).Association("Flavors")
	if fe.FlavorPart == "IMAGE" {
		// Image can only have 1 flavor that has FlavorPart == IMAGE,
		// this is a very bad way of setting contraints through the application layer, and should be refactored.
		for _, f := range ie.Flavors {
			if f.FlavorPart == "IMAGE" {
				if err := assoc.Delete(&fe).Error; err != nil {
					tx.Rollback()
					return err
				}
				break
			}
		}
	}
	if err := assoc.Append(&flavorEntity{ID: flavorUUID}).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "repository/postgres/image_repository:AddAssociatedFlavor() Failed to associate flavor with image")
	}
	return tx.Commit().Error
}

func (repo imageRepo) DeleteByUUID(uuid string) error {
	log.Trace("repository/postgres/image_repository:DeleteByUUID() Entering")
	defer log.Trace("repository/postgres/image_repository:DeleteByUUID() Leaving")

	return repo.db.Delete(imageEntity{}, "id = ?", uuid).Error
}

func (repo imageRepo) DeleteAssociatedFlavor(imageUUID string, flavorUUID string) error {
	log.Trace("repository/postgres/image_repository:DeleteAssociatedFlavor() Entering")
	defer log.Trace("repository/postgres/image_repository:DeleteAssociatedFlavor() Leaving")

	return repo.db.Model(&imageEntity{ID: imageUUID}).Association("Flavors").Delete(&flavorEntity{ID: flavorUUID}).Error
}
