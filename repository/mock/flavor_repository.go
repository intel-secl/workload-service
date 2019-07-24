package mock

import (
	"intel/isecl/workload-service/model"
	flavorUtil "intel/isecl/lib/flavor/util"
	"intel/isecl/workload-service/repository"
)

type MockFlavor struct {
	CreateFn                   func(*flavorUtil.SignedImageFlavor) error
	RetrieveByFilterCriteriaFn func(repository.FlavorFilter) ([]model.Flavor, error)
	RetrieveByUUIDFn           func(string) (*model.Flavor, error)
	RetrieveByLabelFn          func(string) (*model.Flavor, error)
	DeleteFn                   func(*model.Flavor) error
	DeleteByUUIDFn             func(string) error
}

func (m *MockFlavor) Create(f *flavorUtil.SignedImageFlavor) error {
	if m.CreateFn != nil {
		return m.Create(f)
	}
	return nil
}

func (m *MockFlavor) RetrieveByFilterCriteria(locator repository.FlavorFilter) ([]model.Flavor, error) {
	if m.RetrieveByFilterCriteriaFn != nil {
		return m.RetrieveByFilterCriteriaFn(locator)
	}
	flav := f
	flav.Image.Meta.Description.Label = locator.Label
	return []model.Flavor{flav}, nil
}

func (m *MockFlavor) RetrieveByUUID(uuid string) (*model.Flavor, error) {
	if m.RetrieveByUUIDFn != nil {
		return m.RetrieveByUUIDFn(uuid)
	}
	flav := f
	flav.Image.Meta.ID = uuid
	return &flav, nil
}

func (m *MockFlavor) RetrieveByLabel(label string) (*model.Flavor, error) {
	if m.RetrieveByLabelFn != nil {
		return m.RetrieveByLabelFn(label)
	}
	flav := f
	flav.Image.Meta.Description.Label = label
	return &flav, nil
}

func (m *MockFlavor) Delete(f *model.Flavor) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(f)
	}
	return nil
}

func (m *MockFlavor) DeleteByUUID(u string) error {
	if m.DeleteByUUIDFn != nil {
		return m.DeleteByUUIDFn(u)
	}
	return nil
}
