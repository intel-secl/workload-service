package postgres

import (
	"encoding/json"
	"errors"
	flvr "intel/isecl/lib/flavor"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type flavorEntity struct {
	ID         string `gorm:"type:uuid;primary_key;"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Label      string         `gorm:"unique;not null"`
	FlavorPart string         `gorm:"not null"`
	Content    postgres.Jsonb `gorm:"type:jsonb;not null"`
	Signature  string
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

func (fe *flavorEntity) unmarshal() (*flvr.SignedImageFlavor, error) {
	var i flvr.SignedImageFlavor
	// ignore error since we validate it on callbacks
	err := json.Unmarshal(fe.Content.RawMessage, &i.ImageFlavor)
	i.Signature = fe.Signature
	return &i, err
}

func (fe *flavorEntity) Flavor() flvr.SignedImageFlavor {
	i, _ := fe.unmarshal()
	return *i
}
