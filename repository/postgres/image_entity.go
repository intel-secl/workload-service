/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package postgres

import (
	"intel/isecl/workload-service/model"
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
	return "images"
}

func (ie *imageEntity) Image() model.Image {
	flavorIDs := make([]string, len(ie.Flavors))
	for i, fe := range ie.Flavors {
		flavorIDs[i] = fe.ID
	}
	return model.Image{
		ID:        ie.ID,
		FlavorIDs: flavorIDs,
	}
}
