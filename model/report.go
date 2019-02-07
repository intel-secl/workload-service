package model

import (
	"intel/isecl/lib/verifier"
	"intel/isecl/lib/common/crypt"
)

// Report is an alias to verifier.VMTrustReport
type Report struct {
	ID string `json:"id,omitempty"`
	verifier.VMTrustReport
	crypt.SignedData
}
