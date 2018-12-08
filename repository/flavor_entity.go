package repository

import (
	"encoding/json"
	"errors"
	"intel/isecl/lib/flavor"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type flavorEntity struct {
	// alias gorm.Model
	ID        string `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// DeletedAt *time.Time
	// Above 4 are Aliases
	Label               string         `gorm:"unique"`
	Content             postgres.Jsonb `gorm:"type:jsonb;not null"`
	*flavor.ImageFlavor `gorm:"-"`
}

// BeforeCreate coerces FlavorEntity Primary Key's (ID) to be a UUID instead of integer
func (fe *flavorEntity) BeforeCreate(scope *gorm.Scope) error {
	if fe.ImageFlavor == nil {
		return errors.New("Content must not be null")
	}
	// none of the below can be nil, as they are not pointers
	id := fe.Image.Meta.ID
	label := fe.Image.Meta.Description.Label
	if len(id) == 0 {
		return errors.New("flavor uuid cannot be empty")
	}
	jsonData, err := json.Marshal(fe)
	if err != nil {
		return err
	}
	if err := scope.SetColumn("id", id); err != nil {
		return err
	}
	if len(label) == 0 {
		if err := scope.SetColumn("label", nil); err != nil {
			return err
		}
	} else {
		if err := scope.SetColumn("label", label); err != nil {
			return err
		}
	}
	if err := scope.SetColumn("content", jsonData); err != nil {
		return err
	}
	return nil
}

func (fe *flavorEntity) AfterFind(scope *gorm.Scope) error {
	// unmarshal Content into ImageFlavor
	var f flavor.ImageFlavor
	err := json.Unmarshal(fe.Content.RawMessage, &f)
	if err != nil {
		return err
	}
	fe.ImageFlavor = &f
	return nil
}
