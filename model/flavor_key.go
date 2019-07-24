package model

import "intel/isecl/lib/flavor"

// FlavorKey is the output for the RPC call to /images/{id}/flavor-key
type FlavorKey struct {
	Flavor    flavor.Image `json:"flavor"`
	Signature string       `json:"signature"`
	Key       []byte       `json:"key"`
}
