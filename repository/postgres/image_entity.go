/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package postgres

import (
	"intel/isecl/workload-service/v2/model"
	"time"
)

type imageEntity struct {
	ID        string `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// Gorm many to many
	Flavors []flavorEntity `gorm:"many2many:image_flavors;association_jointable_foreignkey:flavor_id;jointable_foreignkey:image_id;association_autoupdate:false;association_autocreate:false"`
}

func (ie imageEntity) TableName() string {
	log.Trace("repository/postgres/image_entity:TableName() Entering")
	defer log.Trace("repository/postgres/image_entity:TableName() Leaving")
	return "images"
}

func (ie *imageEntity) Image() model.Image {
	log.Trace("repository/postgres/image_entity:Image() Entering")
	defer log.Trace("repository/postgres/image_entity:Image() Leaving")

	flavorIDs := make([]string, len(ie.Flavors))
	for i, fe := range ie.Flavors {
		flavorIDs[i] = fe.ID
	}
	return model.Image{
		ID:        ie.ID,
		FlavorIDs: flavorIDs,
	}
}
