package model

type Image struct {
	ID        string   `json:"image_id"`
	FlavorIDs []string `json:"flavor_ids,omitempty"`
}
