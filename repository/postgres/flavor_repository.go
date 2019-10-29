/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package postgres

import (
	"encoding/json"
	"github.com/pkg/errors"
	flvr "intel/isecl/lib/flavor"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type flavorRepo struct {
	db *gorm.DB
}

func (repo flavorRepo) Create(f *flvr.SignedImageFlavor) error {
	log.Trace("repository/postgres/flavor_repository:Create() Entering")
	defer log.Trace("repository/postgres/flavor_repository:Create() Leaving")

	log.Debug("repository/postgres/flavor_repository:Create() Create Image flavor")
	if f == nil {
		return errors.New("repository/postgres/flavor_repository:Create() cannot create nil flavor")
	}
	tx := repo.db.Begin()
	var fe flavorEntity
	if !tx.Where("id = ?", f.ImageFlavor.Meta.ID).Or("label = ?", f.ImageFlavor.Meta.Description.Label).Take(&fe).RecordNotFound() {
		// duplicate exists
		tx.Rollback()
		if fe.ID == f.ImageFlavor.Meta.ID {
			return repository.ErrFlavorUUIDAlreadyExists
		} else if fe.Label == f.ImageFlavor.Meta.Description.Label {
			return repository.ErrFlavorLabelAlreadyExists
		}
	}
	flavorJSON, err := json.Marshal(f.ImageFlavor)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "repository/postgres/flavor_repository:Create() failed to marshal ImageFlavor to JSON")
	}
	if err := tx.Create(&flavorEntity{
		ID:         f.ImageFlavor.Meta.ID,
		Label:      f.ImageFlavor.Meta.Description.Label,
		FlavorPart: f.ImageFlavor.Meta.Description.FlavorPart,
		Content:    postgres.Jsonb{RawMessage: flavorJSON},
		Signature:  f.Signature,
	}).Error; err != nil {
		tx.Rollback()
		return errors.Wrap(err, "repository/postgres/flavor_repository:Create() Error in flavor creation")
	}
	return tx.Commit().Error
}
func (repo flavorRepo) RetrieveByFilterCriteria(filter repository.FlavorFilter) ([]model.Flavor, error) {
	log.Trace("repository/postgres/flavor_repository:RetrieveByFilterCriteria() Entering")
	defer log.Trace("repository/postgres/flavor_repository:RetrieveByFilterCriteria() Leaving")
	var flavorEntities []flavorEntity

	log.Debug("repository/postgres/flavor_repository:RetrieveByFilterCriteria() Retrieve flavor by filter criteria")
	if len(filter.FlavorID) > 0 {
		repo.db.Where("id = ?", filter.FlavorID).Find(&flavorEntities)
		return getFlavorModels(flavorEntities)
	}

	if len(filter.Label) > 0 {
		repo.db.Where("label = ?", filter.Label).Find(&flavorEntities)
		return getFlavorModels(flavorEntities)
	}

	if !filter.Filter {
		repo.db.Find(&flavorEntities)
		return getFlavorModels(flavorEntities)
	}

	return nil, errors.New("repository/postgres/flavor_repository:RetrieveByFilterCriteria() invalid flavor filter criteria")
}

func getFlavorModels(flavorEntities []flavorEntity) ([]model.Flavor, error) {
	log.Trace("repository/postgres/flavor_repository:getFlavorModels() Entering")
	defer log.Trace("repository/postgres/flavor_repository:getFlavorModels() Leaving")

	log.Debug("repository/postgres/flavor_repository:getFlavorModels() Retrieve flavor models")
	flavors := make([]model.Flavor, len(flavorEntities))
	for i, v := range flavorEntities {
		flavors[i].Image = v.Flavor().ImageFlavor
	}
	return flavors, nil
}

func (repo flavorRepo) RetrieveByUUID(uuid string) (*model.Flavor, error) {
	log.Trace("repository/postgres/flavor_repository:RetrieveByUUID() Entering")
	defer log.Trace("repository/postgres/flavor_repository:RetrieveByUUID() Leaving")

	log.Debug("repository/postgres/flavor_repository:RetrieveByUUID() Retrieve flavor by UUID")
	var fe flavorEntity
	var f model.Flavor
	fe.ID = uuid
	if err := repo.db.First(&fe).Error; err != nil {
		return nil, errors.Wrap(err, "repository/postgres/flavor_repository:RetrieveByUUID() Failed to retrieve flavor by UUID")
	}
	f.Image = fe.Flavor().ImageFlavor
	return &f, nil
}

func (repo flavorRepo) RetrieveByLabel(label string) (*model.Flavor, error) {
	log.Trace("repository/postgres/flavor_repository:RetrieveByLabel() Entering")
	defer log.Trace("repository/postgres/flavor_repository:RetrieveByLabel() Leaving")

	log.Debug("repository/postgres/flavor_repository:RetrieveByLabel() Retrieve flavor by label")
	var fe flavorEntity
	var f model.Flavor
	if err := repo.db.Where("label = ?", label).Find(&fe).Error; err != nil {
		return nil, errors.Wrap(err, "repository/postgres/flavor_repository:RetrieveByLabel() Failed to retrieve flavor by Label")
	}
	f.Image = fe.Flavor().ImageFlavor
	return &f, nil
}

func (repo flavorRepo) Delete(f *model.Flavor) error {
	log.Trace("repository/postgres/flavor_repository:Delete() Entering")
	defer log.Trace("repository/postgres/flavor_repository:Delete() Leaving")

	log.Debug("repository/postgres/flavor_repository:Delete() Delete flavor")
	if f == nil {
		return errors.New("repository/postgres/flavor_repository:Delete() cannot delete nil flavor")
	}
	return repo.DeleteByUUID(f.Image.Meta.ID)
}

func (repo flavorRepo) DeleteByUUID(uuid string) error {
	log.Trace("repository/postgres/flavor_repository:DeleteByUUID() Entering")
	defer log.Trace("repository/postgres/flavor_repository:DeleteByUUID() Leaving")

	log.Debug("repository/postgres/flavor_repository:DeleteByUUID() Delete flavor by UUID")
	return repo.db.Delete(&flavorEntity{ID: uuid}).Error
	// Delete associated images
	//return repo.db.Where("flavor_id = ?", uuid).Delete(imageEntity{}).Error
}
