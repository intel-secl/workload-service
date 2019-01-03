package postgres

import (
	"encoding/json"
	"errors"
	"intel/isecl/workload-service/model"
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
	Label      string         `gorm:"unique;not null"`
	FlavorPart string         `gorm:"not null"`
	Content    postgres.Jsonb `gorm:"type:jsonb;not null"`
	//Images  []imageEntity  `gorm:"many2many:image_flavors"`
}

func (fe flavorEntity) TableName() string {
	return "flavors"
}

func (fe *flavorEntity) BeforeCreate(scope *gorm.Scope) error {
	if !json.Valid(fe.Content.RawMessage) {
		return errors.New("JSON Content is not valid")
	}
	// try and unmarshal it
	_, err := fe.unmarshal()
	if err != nil {
		return errors.New("JSON Content does not match flavor schema")
	}
	return nil
}

func (fe *flavorEntity) AfterFind(scope *gorm.Scope) error {
	// try and unmarshal it
	_, err := fe.unmarshal()
	if err != nil {
		return errors.New("JSON Content does not match flavor schema")
	}
	return nil
}

func (fe *flavorEntity) unmarshal() (*model.Flavor, error) {
	var i model.Flavor
	// ignore error since we validate it on callbacks
	err := json.Unmarshal(fe.Content.RawMessage, &i)
	return &i, err
}

func (fe *flavorEntity) Flavor() model.Flavor {
	i, _ := fe.unmarshal()
	return *i
}
