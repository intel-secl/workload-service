package model

type Image struct {
	ID       string `gorm:"type:uuid;primary_key;" json:"image_id"`
	FlavorID string `gorm:"type:uuid;not null" json:"flavor_id"`
}
