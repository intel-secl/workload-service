/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package mock

import (
	flvr "intel/isecl/lib/flavor/v4"
	"intel/isecl/workload-service/v4/model"
	"intel/isecl/workload-service/v4/repository"
)

type MockFlavor struct {
	CreateFn                   func(*flvr.SignedImageFlavor) error
	RetrieveByFilterCriteriaFn func(repository.FlavorFilter) ([]model.Flavor, error)
	RetrieveByUUIDFn           func(string) (*model.Flavor, error)
	RetrieveByLabelFn          func(string) (*model.Flavor, error)
	DeleteFn                   func(*model.Flavor) error
	DeleteByUUIDFn             func(string) error
}

func (m *MockFlavor) Create(f *flvr.SignedImageFlavor) error {
	log.Trace("repository/mock/flavor_repository:Create() Entering")
	defer log.Trace("repository/mock/flavor_repository:Create() Leaving")
	log.Debug("repository/mock/flavor_repository:Create() Create mock image flavor")
	if m.CreateFn != nil {
		return m.Create(f)
	}
	return nil
}

func (m *MockFlavor) RetrieveByFilterCriteria(locator repository.FlavorFilter) ([]model.Flavor, error) {
	log.Trace("repository/mock/flavor_repository:RetrieveByFilterCriteria() Entering")
	defer log.Trace("repository/mock/flavor_repository:RetrieveByFilterCriteria() Leaving")
	log.Debug("repository/mock/flavor_repository:RetrieveByFilterCriteria() Retrieve mock image flavor by filter criteria")
	if m.RetrieveByFilterCriteriaFn != nil {
		return m.RetrieveByFilterCriteriaFn(locator)
	}
	flav := f
	flav.Image.Meta.Description.Label = locator.Label
	return []model.Flavor{flav}, nil
}

func (m *MockFlavor) RetrieveByUUID(uuid string) (*model.Flavor, error) {
	log.Trace("repository/mock/flavor_repository:RetrieveByUUID() Entering")
	defer log.Trace("repository/mock/flavor_repository:RetrieveByUUID() Leaving")
	log.Debug("repository/mock/flavor_repository:RetrieveByUUID() Retrieve mock image flavor by UUID")
	if m.RetrieveByUUIDFn != nil {
		return m.RetrieveByUUIDFn(uuid)
	}
	flav := f
	flav.Image.Meta.ID = uuid
	return &flav, nil
}

func (m *MockFlavor) RetrieveByLabel(label string) (*model.Flavor, error) {
	log.Trace("repository/mock/flavor_repository:RetrieveByLabel() Entering")
	defer log.Trace("repository/mock/flavor_repository:RetrieveByLabel() Leaving")
	log.Debug("repository/mock/flavor_repository:RetrieveByLabel() Retrieve mock image flavor by Label")
	if m.RetrieveByLabelFn != nil {
		return m.RetrieveByLabelFn(label)
	}
	flav := f
	flav.Image.Meta.Description.Label = label
	return &flav, nil
}

func (m *MockFlavor) Delete(f *model.Flavor) error {
	log.Trace("repository/mock/flavor_repository:Delete() Entering")
	defer log.Trace("repository/mock/flavor_repository:Delete() Leaving")
	log.Debug("repository/mock/flavor_repository:Delete() Delete mock image flavor")
	if m.DeleteFn != nil {
		return m.DeleteFn(f)
	}
	return nil
}

func (m *MockFlavor) DeleteByUUID(u string) error {
	log.Trace("repository/mock/flavor_repository:DeleteByUUID() Entering")
	defer log.Trace("repository/mock/flavor_repository:DeleteByUUID() Leaving")
	log.Debug("repository/mock/flavor_repository:DeleteByUUID() Delete mock image flavor by UUID")
	if m.DeleteByUUIDFn != nil {
		return m.DeleteByUUIDFn(u)
	}
	return nil
}
