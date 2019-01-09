package postgres

import (
	"encoding/json"
	"errors"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"

	"github.com/jinzhu/gorm/dialects/postgres"

	"github.com/jinzhu/gorm"
)

type flavorRepo struct {
	db *gorm.DB
}

func (repo flavorRepo) Create(f *model.Flavor) error {
	if f == nil {
		return errors.New("cannot create nil flavor")
	}
	tx := repo.db.Begin()
	var fe flavorEntity
	if !tx.Where("id = ?", f.Image.Meta.ID).Or("label = ?", f.Image.Meta.Description.Label).Take(&fe).RecordNotFound() {
		// duplicate exists
		tx.Rollback()
		if fe.ID == f.Image.Meta.ID {
			return repository.ErrFlavorUUIDAlreadyExists
		} else if fe.Label == f.Image.Meta.Description.Label {
			return repository.ErrFlavorLabelAlreadyExists
		} else {
			// panic since this shoudln't be reached, indicates a critical error in the code/db logic
			panic("This shouldn't be reached, logic error")
		}
	}
	flavorJSON, err := json.Marshal(f)
	if err != nil {
		return errors.New("failed to marshal ImageFlavor to JSON")
	}
	if err := tx.Create(&flavorEntity{
		ID:         f.Image.Meta.ID,
		Label:      f.Image.Meta.Description.Label,
		FlavorPart: f.Image.Meta.Description.FlavorPart,
		Content:    postgres.Jsonb{RawMessage: flavorJSON},
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
func (repo flavorRepo) RetrieveByFilterCriteria(filter repository.FlavorFilter) ([]model.Flavor, error) {
	var entities []flavorEntity
	if err := repo.db.Find(&entities).Error; err != nil {
		return nil, err
	}
	flavors := make([]model.Flavor, len(entities))
	for i, v := range entities {
		flavors[i] = v.Flavor()
	}
	return flavors, nil
}

func (repo flavorRepo) RetrieveByUUID(uuid string) (*model.Flavor, error) {
	var fe flavorEntity
	fe.ID = uuid
	if err := repo.db.First(&fe).Error; err != nil {
		return nil, err
	}
	f := fe.Flavor()
	return &f, nil
}

func (repo flavorRepo) RetrieveByLabel(label string) (*model.Flavor, error) {
	var fe flavorEntity
	if err := repo.db.Where("label = ?", label).Find(&fe).Error; err != nil {
		return nil, err
	}
	f := fe.Flavor()
	return &f, nil
}

func (repo flavorRepo) Delete(f *model.Flavor) error {
	if f == nil {
		return errors.New("cannot delete nil flavor")
	}
	return repo.DeleteByUUID(f.Image.Meta.ID)
}

func (repo flavorRepo) DeleteByUUID(uuid string) error {
	return repo.db.Delete(&flavorEntity{ID: uuid}).Error
	// Delete associated images
	//return repo.db.Where("flavor_id = ?", uuid).Delete(imageEntity{}).Error
}
