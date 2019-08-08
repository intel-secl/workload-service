package mock

import (
	flvr "intel/isecl/lib/flavor"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
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
	if m.CreateFn != nil {
		return m.CreateFn(image)
	}
	return nil
}

func (m *MockImage) RetrieveByUUID(uuid string) (*model.Image, error) {
	if m.RetrieveByUUIDFn != nil {
		return m.RetrieveByUUIDFn(uuid)
	}
	image := i
	image.ID = uuid
	return &i, nil
}

func (m *MockImage) RetrieveAssociatedImageFlavor(imageUUID string) (*flvr.SignedImageFlavor, error) {
	if m.RetrieveAssociatedImageFlavorFn != nil {
		return m.RetrieveAssociatedImageFlavorFn(imageUUID)
	}
	return &signedFlavor, nil
}

func (m *MockImage) RetrieveByFilterCriteria(locator repository.ImageFilter) ([]model.Image, error) {
	if m.RetrieveByFilterCriteriaFn != nil {
		return m.RetrieveByFilterCriteriaFn(locator)
	}
	return []model.Image{i}, nil
}

func (m *MockImage) RetrieveAssociatedFlavor(imageUUID string, flavorUUID string) (*model.Flavor, error) {
	if m.RetrieveAssociatedFlavorFn != nil {
		return m.RetrieveAssociatedFlavorFn(imageUUID, flavorUUID)
	}
	return &f, nil
}

func (m *MockImage) RetrieveAssociatedFlavorByFlavorPart(imageUUID string, flavorPart string) (*flvr.SignedImageFlavor, error) {
	if m.RetrieveAssociatedFlavorFn != nil {
		return m.RetrieveAssociatedFlavorByFlavorPartFn(imageUUID, flavorPart)
	}
	return &signedFlavor, nil
}

func (m *MockImage) RetrieveAssociatedFlavors(imageUUID string) ([]model.Flavor, error) {
	if m.RetrieveAssociatedFlavorsFn != nil {
		return m.RetrieveAssociatedFlavorsFn(imageUUID)
	}
	return []model.Flavor{f}, nil
}

func (m *MockImage) Update(image *model.Image) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(image)
	}
	return nil
}

func (m *MockImage) AddAssociatedFlavor(imageID string, flavorID string) error {
	if m.AddAssociatedFlavorFn != nil {
		return m.AddAssociatedFlavorFn(imageID, flavorID)
	}
	return nil
}

func (m *MockImage) DeleteByUUID(imageID string) error {
	if m.DeleteByUUIDFn != nil {
		return m.DeleteByUUIDFn(imageID)
	}
	return nil
}

func (m *MockImage) DeleteAssociatedFlavor(imageID string, flavorID string) error {
	if m.DeleteAssociatedFlavorFn != nil {
		return m.DeleteAssociatedFlavorFn(imageID, flavorID)
	}
	return nil
}
