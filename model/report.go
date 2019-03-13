package model

import (
	"intel/isecl/lib/common/crypt"
	"intel/isecl/lib/verifier"
)

// Report is an alias to verifier.VMTrustReport
type Report struct {
	ID string `json:"id,omitempty"`
	verifier.InstanceTrustReport
	crypt.SignedData
}
