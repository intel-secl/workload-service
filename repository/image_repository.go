package repository

import (
	"errors"
	"intel/isecl/workload-service/model"
)

var (
	ErrImageAssociationAlreadyExists        = errors.New("image association with UUID already exists")
	ErrImageAssociationFlavorDoesNotExist   = errors.New("one or more FlavorID's does not exist in the database")
	ErrImageAssociationDuplicateFlavor      = errors.New("flavor with UUID already associated with image")
	ErrImageAssociationDuplicateImageFlavor = errors.New("image can only be associated with one flavor with FlavorPart = IMAGE")
)

// ImageFilter specifies query filter criteria for retrieving images. Each Field may be empty
type ImageFilter struct {
	FlavorID string `json:"flavor_id,omitempty"`
}

// ImageRepository defines an interface that provides persistence operations for an Image-Flavor link.
// It defines High Level CRUD operations that could be implemented by any database or persistence layer (such as postgres)
// The CRUD operations are logically grouped, but not defined to any single interface, so that FlavorRepository may customize them to its own needs, with
// Stronger typing rather than cast everything from an interface{}
type ImageRepository interface {
	// C
	Create(image *model.Image) error
	// R
	RetrieveByUUID(uuid string) (*model.Image, error)
	RetrieveAssociatedImageFlavor(imageUUID string) (*model.Flavor, error)
	RetrieveAssociatedFlavor(imageUUID string, flavorUUID string) (*model.Flavor, error)
	RetrieveAssociatedFlavors(uuid string) ([]model.Flavor, error)
	RetrieveByFilterCriteria(locator ImageFilter) ([]model.Image, error)
	// U
	Update(image *model.Image) error
	AddAssociatedFlavor(imageUUID string, flavorUUID string) error
	// D
	DeleteByUUID(uuid string) error
	DeleteAssociatedFlavor(imageUUID string, flavorUUID string) error
}
