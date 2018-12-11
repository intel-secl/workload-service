package repository

import (
	"time"
)

type imageEntity struct {
	// alias gorm.Model
	ID        string `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	// Gorm Belongs-To association
	Flavor   flavorEntity `gorm:"foreignKey:FlavorID"`
	FlavorID string       `gorm:"not null"`
}
