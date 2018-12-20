package repository

import (
	"errors"
	"intel/isecl/lib/flavor"
)

var (
	ErrFlavorUUIDAlreadyExists  = errors.New("flavor already exists with UUID")
	ErrFlavorLabelAlreadyExists = errors.New("flavor already exists with label")
)

// FlavorRepository defines an interface that provides persistence operations for a Flavor.
// It defines High Level CRUD operations that could be implemented by any database or persistence layer (such as postgres)
// The CRUD operations are logically grouped, but not defined to any single interface, so that FlavorRepository may customize them to its own needs, with
// Stronger typing rather than cast everything from an interface{}
type FlavorRepository interface {
	// C
	Create(f *flavor.ImageFlavor) error
	// R
	RetrieveByUUID(uuid string) (*flavor.ImageFlavor, error)
	RetrieveByLabel(label string) (*flavor.ImageFlavor, error)
	// D
	Delete(f *flavor.ImageFlavor) error
	DeleteByUUID(uuid string) error
	DeleteByLabel(label string) error
}
