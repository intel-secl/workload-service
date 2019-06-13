package repository

import (
	"errors"
	"intel/isecl/workload-service/model"
)

var (
	// ErrFlavorUUIDAlreadyExists error when flavor with same UUID already exists in the database
	ErrFlavorUUIDAlreadyExists = errors.New("flavor already exists with UUID")
	// ErrFlavorLabelAlreadyExists error when flavor with same label name already exists in the database
	ErrFlavorLabelAlreadyExists = errors.New("flavor already exists with label")
)

// FlavorFilter defines filter criteria for searching
type FlavorFilter struct {
	FlavorID string `json:"id,omitempty"`
	Label    string `json:"label,omitempty"`
	Filter   bool   `json:"filter,omitempty"`
}

// FlavorRepository defines an interface that provides persistence operations for a Flavor.
// It defines High Level CRUD operations that could be implemented by any database or persistence layer (such as postgres)
// The CRUD operations are logically grouped, but not defined to any single interface, so that FlavorRepository may customize them to its own needs, with
// Stronger typing rather than cast everything from an interface{}
type FlavorRepository interface {
	// C
	Create(f *model.Flavor) error
	// R
	RetrieveByFilterCriteria(filter FlavorFilter) ([]model.Flavor, error)
	RetrieveByUUID(uuid string) (*model.Flavor, error)
	RetrieveByLabel(label string) (*model.Flavor, error)
	// D
	Delete(f *model.Flavor) error
	DeleteByUUID(uuid string) error
}
