package repository

import (
	"errors"
	"intel/isecl/lib/flavor"

	"github.com/jinzhu/gorm"
	// Import Postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	ErrFlavorUUIDAlreadyExists  = errors.New("flavor already exists with UUUID")
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

type flavorPsql struct {
	db *gorm.DB
}

func (repo *flavorPsql) Create(f *flavor.ImageFlavor) error {
	if f == nil {
		return errors.New("cannot create nil flavor")
	}
	tx := repo.db.Begin()
	var fe flavorEntity
	if !tx.Where("id = ? OR label = ?", f.Image.Meta.ID, f.Image.Meta.Description.Label).Take(&fe).RecordNotFound() {
		// duplicate exists
		tx.Rollback()
		if fe.ID == f.Image.Meta.ID {
			return ErrFlavorUUIDAlreadyExists
		} else if fe.Label == f.Image.Meta.Description.Label {
			return ErrFlavorLabelAlreadyExists
		} else {
			// panic since this shouldn't be reached, indicates a critical error in the code/db logic
			panic("This shouldn't be reached, logic error")
		}
	}
	if err := tx.Create(&flavorEntity{ImageFlavor: f}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (repo *flavorPsql) RetrieveByUUID(uuid string) (*flavor.ImageFlavor, error) {
	var fe flavorEntity
	fe.ID = uuid
	if err := repo.db.First(&fe).Error; err != nil {
		return nil, err
	}
	return fe.ImageFlavor, nil
}

func (repo *flavorPsql) RetrieveByLabel(label string) (*flavor.ImageFlavor, error) {
	var fe flavorEntity
	if err := repo.db.Where("label = ?", label).Find(&fe).Error; err != nil {
		return nil, err
	}
	return fe.ImageFlavor, nil
}

func (repo *flavorPsql) Delete(f *flavor.ImageFlavor) error {
	if f == nil {
		return errors.New("cannot delete nil flavor")
	}
	return repo.DeleteByUUID(f.Image.Meta.ID)
}

func (repo *flavorPsql) DeleteByUUID(uuid string) error {
	return repo.db.Delete(&flavorEntity{ID: uuid}).Error
}

func (repo *flavorPsql) DeleteByLabel(label string) error {
	return repo.db.Where("label = ?", label).Delete(&flavorEntity{}).Error
}

var singleton *flavorPsql

// GetFlavorRepository gets the global instance of the FlavorRepository. Currently is only backed by postgresql
func GetFlavorRepository() (FlavorRepository, error) {
	if singleton == nil {
		// try to open gorm
		db, err := gorm.Open("postgres", "host=localhost port=5432 user=postgres dbname=postgres password=test sslmode=disable")
		if err != nil {
			return nil, err
		}
		db.AutoMigrate(&flavorEntity{})
		singleton = &flavorPsql{
			db: db,
		}
	}
	return singleton, nil
}
