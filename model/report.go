/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package model

import (
	"intel/isecl/lib/common/v4/crypt"
	"intel/isecl/lib/verifier/v4"
)

// Report is an alias to verifier.VMTrustReport
type Report struct {
	ID string `json:"id,omitempty"`
	verifier.InstanceTrustReport
	crypt.SignedData
}
