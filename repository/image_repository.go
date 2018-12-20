package repository

import (
	"errors"
	"intel/isecl/workload-service/model"
)

var (
	ErrImageAssociationAlreadyExists = errors.New("image association with UUID already exists")
)

// ImageLocator specifies query filter criteria for retrieving images. Each Field may be empty
type ImageLocator struct {
	FlavorID string `json:"flavor_id, omitempty"`
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
	RetrieveByFilterCriteria(locator ImageLocator) ([]model.Image, error)
	// U
	// D
	DeleteByUUID(uuid string) error
}
