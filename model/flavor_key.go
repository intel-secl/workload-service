package model

// FlavorKey is the output for the RPC call to /images/{id}/flavor-key
type FlavorKey struct {
	Flavor `json: flavor`
	Key    []byte `json: key`
}
