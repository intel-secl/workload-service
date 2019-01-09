package model

import (
	"encoding/json"
	"intel/isecl/lib/flavor"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FlavorKeyInfo struct {
	flavor.ImageFlavor
	Key []byte `json:"key"`
}

func TestFlavorKeyDeserialize(t *testing.T) {
	raw := `{
		"flavor": {
		  "meta": {
			"id": "c10cc1f5-7c32-4d24-9a1e-17cb5f6b125d",
			"description": {
			  "flavor_part": "IMAGE",
			  "label": "label"
			}
		  },
		  "encryption": {
			"encryption_required": true,
			"key_url": "https://10.1.68.68:443/v1/keys/ecee021e-9669-4e53-9224-8880fb4e4080/transfer",
			"digest": "C6CR+CsFIqfxEgkoe4hWVQyBwJ71amh1TwVQBo4TWa0="
		  }
		},
		"key": "6WNBU9HfOF0hVynmGaTOG3wSc4R0hlxWe0hOXW3WAwNwjpiyo5/OExrfYkIhHwRlBWPhkpKfAy6gJ9AOzPr3VJFCuU+chtBNtPcgoXIxa02vJSiLGmYOXzaVDUZD1Ht7ND2rrVQB8cqjYC/0FXfpUu65/KFhbvTE1liXrmfmfFbrgQGoDqIA8rjuMtZTTLDuIZiOQZR9jdhjysDzEkZUDglqIbN+fIY/8c1Bsge7E26YEpBWqPNTwmuNmFwHRr4f6HSGmbUwVUietwGjRkbp/CSMIiWyLI99hvuQ/D5Tc1jEN4wvGPTvzc9MpecJ2at0BHRTx4MpEQ+u3XWW3zy1xA=="
	  }`
	var fk FlavorKeyInfo
	json.Unmarshal([]byte(raw), &fk)
	assert.Equal(t, "c10cc1f5-7c32-4d24-9a1e-17cb5f6b125d", fk.Image.Meta.ID)
}
