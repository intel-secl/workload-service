/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package postgres

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
	commLog "intel/isecl/lib/common/log"
	flvr "intel/isecl/lib/flavor"
	"time"
)

var log = commLog.GetDefaultLogger()

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
	log.Trace("repository/postgres/flavor_entity:TableName() Entering")
	defer log.Trace("repository/postgres/flavor_entity:TableName() Leaving")
	return "flavors"
}

func (fe *flavorEntity) BeforeCreate(scope *gorm.Scope) error {
	log.Trace("repository/postgres/flavor_entity:BeforeCreate() Entering")
	defer log.Trace("repository/postgres/flavor_entity:BeforeCreate() Leaving")

	if !json.Valid(fe.Content.RawMessage) {
		return errors.New("repository/postgres/flavor_entity:BeforeCreate() JSON Content is not valid")
	}
	// try and unmarshal it
	_, err := fe.unmarshal()
	if err != nil {
		return errors.Wrap(err, "repository/postgres/flavor_entity:BeforeCreate() JSON Content does not match flavor schema")
	}
	return nil
}

func (fe *flavorEntity) AfterFind(scope *gorm.Scope) error {
	log.Trace("repository/postgres/flavor_entity:AfterFind() Entering")
	defer log.Trace("repository/postgres/flavor_entity:AfterFind() Leaving")

	// try and unmarshal it
	_, err := fe.unmarshal()
	if err != nil {
		return errors.Wrap(err, "repository/postgres/flavor_entity:AfterFind() JSON Content does not match flavor schema")
	}
	return nil
}

func (fe *flavorEntity) unmarshal() (*flvr.SignedImageFlavor, error) {
	log.Trace("repository/postgres/flavor_entity:unmarshal() Entering")
	defer log.Trace("repository/postgres/flavor_entity:unmarshal() Leaving")

	var i flvr.SignedImageFlavor
	// ignore error since we validate it on callbacks
	err := json.Unmarshal(fe.Content.RawMessage, &i.ImageFlavor)
	i.Signature = fe.Signature
	return &i, errors.Wrap(err, "repository/postgres/flavor_entity:unmarshal() Error in unmarshalling the data")
}

func (fe *flavorEntity) Flavor() flvr.SignedImageFlavor {
	log.Trace("repository/postgres/flavor_entity:Flavor() Entering")
	defer log.Trace("repository/postgres/flavor_entity:Flavor() Leaving")

	i, _ := fe.unmarshal()
	return *i
}
