package model

type Image struct {
	ID        string   `json:"id"`
	FlavorIDs []string `json:"flavor_ids"`
}
