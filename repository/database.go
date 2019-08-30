/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package repository

import (
	"github.com/jinzhu/gorm"
)

type WlsDatabase interface {
	Migrate() error
	FlavorRepository() FlavorRepository
	ImageRepository() ImageRepository
	ReportRepository() ReportRepository
	Driver() *gorm.DB
}
