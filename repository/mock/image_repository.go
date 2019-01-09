package mock

import (
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
)

type mockImage struct{}

func (m mockImage) Create(image *model.Image) error {
	return nil
}

func (m mockImage) RetrieveByUUID(uuid string) (*model.Image, error) {
	image := i
	image.ID = uuid
	return &i, nil
}

func (m mockImage) RetrieveAssociatedImageFlavor(imageUUID string) (*model.Flavor, error) {
	return &f, nil
}

func (m mockImage) RetrieveByFilterCriteria(locator repository.ImageFilter) ([]model.Image, error) {
	return []model.Image{i}, nil
}

func (m mockImage) RetrieveAssociatedFlavor(imageUUID string, flavorUUID string) (*model.Flavor, error) {
	return &f, nil
}

func (m mockImage) RetrieveAssociatedFlavors(imageUUID string) ([]model.Flavor, error) {
	return []model.Flavor{f}, nil
}

func (m mockImage) Update(image *model.Image) error {
	return nil
}

func (m mockImage) AddAssociatedFlavor(string, string) error {
	return nil
}

func (m mockImage) DeleteByUUID(string) error {
	return nil
}

func (m mockImage) DeleteAssociatedFlavor(string, string) error {
	return nil
}
