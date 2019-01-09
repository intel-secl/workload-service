package mock

import (
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
)

type mockFlavor struct {
}

func (m mockFlavor) Create(f *model.Flavor) error {
	return nil
}

func (m mockFlavor) RetrieveByFilterCriteria(locator repository.FlavorFilter) ([]model.Flavor, error) {
	flav := f
	flav.Image.Meta.Description.Label = locator.Label
	return []model.Flavor{flav}, nil
}

func (m mockFlavor) RetrieveByUUID(uuid string) (*model.Flavor, error) {
	flav := f
	flav.Image.Meta.ID = uuid
	return &flav, nil
}

func (m mockFlavor) RetrieveByLabel(label string) (*model.Flavor, error) {
	flav := f
	flav.Image.Meta.Description.Label = label
	return &flav, nil
}

func (m mockFlavor) Delete(*model.Flavor) error {
	return nil
}

func (m mockFlavor) DeleteByUUID(string) error {
	return nil
}
