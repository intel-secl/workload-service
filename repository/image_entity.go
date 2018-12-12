package repository

import (
	"intel/isecl/workload-service/model"
	"time"
)

type imageEntity struct {
	// alias gorm.Model
	model.Image
	//ID        string `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// Gorm Belongs-To association
	Flavor *flavorEntity `gorm:"foreignKey:FlavorID"`
	//FlavorID string       `gorm:"not null"`
}
