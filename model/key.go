/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package model

// RequestKey struct defines input parameters to retrieve a key
type RequestKey struct {
	HwId   string `json:"hardware_uuid"`
	KeyUrl string `json:"key_url"`
}

// ReturnKey to return key Json
type ReturnKey struct {
	Key []byte `json:"key"`
}
