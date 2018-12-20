package model

type Image struct {
	ID        string   `gorm:"type:uuid;not null;" json:"image_id"`
	FlavorIDs []string `gorm:"type:uuid;not null" json:"flavor_ids"`
}
