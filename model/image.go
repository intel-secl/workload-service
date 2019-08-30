/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package model

type Image struct {
	ID        string   `json:"id"`
	FlavorIDs []string `json:"flavor_ids"`
}
