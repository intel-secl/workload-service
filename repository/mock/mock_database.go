/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package mock

import (
	"github.com/jinzhu/gorm"
	"intel/isecl/workload-service/v4/repository"
)

// Database provides a mock Db
type Database struct {
	MockFlavor MockFlavor
	MockImage  MockImage
	MockReport MockReport
}

func (m *Database) Migrate() error {
	log.Trace("repository/mock/mock_database:Migrate() Entering")
	defer log.Trace("repository/mock/mock_database:Migrate() Leaving")
	return nil
}

func (m *Database) FlavorRepository() repository.FlavorRepository {
	log.Trace("repository/mock/mock_database:FlavorRepository() Entering")
	defer log.Trace("repository/mock/mock_database:FlavorRepository() Leaving")
	return &m.MockFlavor
}

func (m *Database) ImageRepository() repository.ImageRepository {
	log.Trace("repository/mock/mock_database:ImageRepository() Entering")
	defer log.Trace("repository/mock/mock_database:ImageRepository() Leaving")
	return &m.MockImage
}

func (m *Database) ReportRepository() repository.ReportRepository {
	log.Trace("repository/mock/mock_database:ReportRepository() Entering")
	defer log.Trace("repository/mock/mock_database:ReportRepository() Leaving")
	return &m.MockReport
}

func (m *Database) Driver() *gorm.DB {
	log.Trace("repository/mock/mock_database:Driver() Entering ")
	defer log.Trace("repository/mock/mock_database:Driver() Leaving")
	return nil
}
