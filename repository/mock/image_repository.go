/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package mock

import (
	flvr "intel/isecl/lib/flavor/v2"
	"intel/isecl/workload-service/v2/model"
	"intel/isecl/workload-service/v2/repository"
)

type MockImage struct {
	CreateFn                               func(*model.Image) error
	RetrieveByUUIDFn                       func(string) (*model.Image, error)
	RetrieveAssociatedImageFlavorFn        func(string) (*flvr.SignedImageFlavor, error)
	RetrieveByFilterCriteriaFn             func(repository.ImageFilter) ([]model.Image, error)
	RetrieveAssociatedFlavorFn             func(string, string) (*model.Flavor, error)
	RetrieveAssociatedFlavorByFlavorPartFn func(string, string) (*flvr.SignedImageFlavor, error)
	RetrieveAssociatedFlavorsFn            func(string) ([]model.Flavor, error)
	UpdateFn                               func(*model.Image) error
	AddAssociatedFlavorFn                  func(string, string) error
	DeleteByUUIDFn                         func(string) error
	DeleteAssociatedFlavorFn               func(string, string) error
}

func (m *MockImage) Create(image *model.Image) error {
	log.Trace("repository/mock/image_repository:Create() Entering")
	defer log.Trace("repository/mock/image_repository:Create() Leaving")
	log.Debug("repository/mock/image_repository:Create() Create mock image")
	if m.CreateFn != nil {
		return m.CreateFn(image)
	}
	return nil
}

func (m *MockImage) RetrieveByUUID(uuid string) (*model.Image, error) {
	log.Trace("repository/mock/image_repository:RetrieveByUUID() Entering")
	defer log.Trace("repository/mock/image_repository:RetrieveByUUID() Leaving")
	log.Debug("repository/mock/image_repository:RetrieveByUUID() Retrieve mock image by UUID")
	if m.RetrieveByUUIDFn != nil {
		return m.RetrieveByUUIDFn(uuid)
	}
	image := i
	image.ID = uuid
	return &i, nil
}

func (m *MockImage) RetrieveAssociatedImageFlavor(imageUUID string) (*flvr.SignedImageFlavor, error) {
	log.Trace("repository/mock/image_repository:RetrieveAssociatedImageFlavor() Entering")
	defer log.Trace("repository/mock/image_repository:RetrieveAssociatedImageFlavor() Leaving")
	log.Debug("repository/mock/image_repository:RetrieveAssociatedImageFlavor() Retrieve associated mock image flavor by image UUID")
	if m.RetrieveAssociatedImageFlavorFn != nil {
		return m.RetrieveAssociatedImageFlavorFn(imageUUID)
	}
	return &signedFlavor, nil
}

func (m *MockImage) RetrieveByFilterCriteria(locator repository.ImageFilter) ([]model.Image, error) {
	log.Trace("repository/mock/image_repository:RetrieveByFilterCriteria() Entering")
	defer log.Trace("repository/mock/image_repository:RetrieveByFilterCriteria() Leaving")
	log.Debug("repository/mock/image_repository:RetrieveByFilterCriteria() Retrieve mock image by filter criteria")
	if m.RetrieveByFilterCriteriaFn != nil {
		return m.RetrieveByFilterCriteriaFn(locator)
	}
	return []model.Image{i}, nil
}

func (m *MockImage) RetrieveAssociatedFlavor(imageUUID string, flavorUUID string) (*model.Flavor, error) {
	log.Trace("repository/mock/image_repository:RetrieveAssociatedFlavor() Entering")
	defer log.Trace("repository/mock/image_repository:RetrieveAssociatedFlavor() Leaving")
	log.Debug("repository/mock/image_repository:RetrieveAssociatedFlavor() Retrieve associated mock image flavor by image UUID and flavor UUID")
	if m.RetrieveAssociatedFlavorFn != nil {
		return m.RetrieveAssociatedFlavorFn(imageUUID, flavorUUID)
	}
	return &f, nil
}

func (m *MockImage) RetrieveAssociatedFlavorByFlavorPart(imageUUID string, flavorPart string) (*flvr.SignedImageFlavor, error) {
	log.Trace("repository/mock/image_repository:RetrieveAssociatedFlavorByFlavorPart() Entering")
	defer log.Trace("repository/mock/image_repository:RetrieveAssociatedFlavorByFlavorPart() Leaving")
	log.Debug("repository/mock/image_repository:RetrieveAssociatedFlavorByFlavorPart() Retrieve associated mock image flavor by flavor part")
	if m.RetrieveAssociatedFlavorFn != nil {
		return m.RetrieveAssociatedFlavorByFlavorPartFn(imageUUID, flavorPart)
	}
	return &signedFlavor, nil
}

func (m *MockImage) RetrieveAssociatedFlavors(imageUUID string) ([]model.Flavor, error) {
	log.Trace("repository/mock/image_repository:RetrieveAssociatedFlavors() Entering")
	defer log.Trace("repository/mock/image_repository:RetrieveAssociatedFlavors() Leaving")
	log.Debug("repository/mock/image_repository:RetrieveAssociatedFlavors() Retrieve associated mock image flavors by image UUID")
	if m.RetrieveAssociatedFlavorsFn != nil {
		return m.RetrieveAssociatedFlavorsFn(imageUUID)
	}
	return []model.Flavor{f}, nil
}

func (m *MockImage) Update(image *model.Image) error {
	log.Trace("repository/mock/image_repository:Update() Entering")
	defer log.Trace("repository/mock/image_repository:Update() Leaving")
	log.Debug("repository/mock/image_repository:Update() Update mock image")
	if m.UpdateFn != nil {
		return m.UpdateFn(image)
	}
	return nil
}

func (m *MockImage) AddAssociatedFlavor(imageID string, flavorID string) error {
	log.Trace("repository/mock/image_repository:AddAssociatedFlavor() Entering")
	defer log.Trace("repository/mock/image_repository:AddAssociatedFlavor() Leaving")
	log.Debug("repository/mock/image_repository:AddAssociatedFlavor() Associate flavor with mock image ")
	if m.AddAssociatedFlavorFn != nil {
		return m.AddAssociatedFlavorFn(imageID, flavorID)
	}
	return nil
}

func (m *MockImage) DeleteByUUID(imageID string) error {
	log.Trace("repository/mock/image_repository:DeleteByUUID() Entering")
	defer log.Trace("repository/mock/image_repository:DeleteByUUID() Leaving")
	log.Debug("repository/mock/image_repository:DeleteByUUID() Delete mock image by UUID")
	if m.DeleteByUUIDFn != nil {
		return m.DeleteByUUIDFn(imageID)
	}
	return nil
}

func (m *MockImage) DeleteAssociatedFlavor(imageID string, flavorID string) error {
	log.Trace("repository/mock/image_repository:DeleteAssociatedFlavor() Entering")
	defer log.Trace("repository/mock/image_repository:DeleteAssociatedFlavor() Leaving")
	log.Debug("repository/mock/image_repository:DeleteAssociatedFlavor() Delete associated mock image flavor")
	if m.DeleteAssociatedFlavorFn != nil {
		return m.DeleteAssociatedFlavorFn(imageID, flavorID)
	}
	return nil
}
