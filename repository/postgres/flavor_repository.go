package postgres

import (
	"encoding/json"
	"errors"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
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
		} 
	}
	flavorJSON, err := json.Marshal(f)
	if err != nil {
		tx.Rollback()
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
	var flavorEntities []flavorEntity

	if len(filter.FlavorID) > 0 {
		repo.db.Where("id = ?", filter.FlavorID).Find(&flavorEntities)
		return getFlavorModels(flavorEntities)
	}

	if len(filter.Label) > 0 {
		repo.db.Where("label = ?", filter.Label).Find(&flavorEntities)
		return getFlavorModels(flavorEntities)
	}

	if !filter.Filter {
		repo.db.Find(&flavorEntities)
		return getFlavorModels(flavorEntities)
	}

	return nil, errors.New("invalid flavor filter criteria")
}

func getFlavorModels(flavorEntities []flavorEntity) ([]model.Flavor, error) {
	flavors := make([]model.Flavor, len(flavorEntities))
	for i, v := range flavorEntities {
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
