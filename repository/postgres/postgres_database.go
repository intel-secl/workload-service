/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package postgres

import (
	"github.com/jinzhu/gorm"
	"intel/isecl/workload-service/repository"
)

type PostgresDatabase struct {
	DB *gorm.DB
}

func (pd PostgresDatabase) Migrate() error {
	log.Trace("repository/postgres/postgres_database:Migrate() Entering")
	defer log.Trace("repository/postgres/postgres_database:Migrate() Leaving")

	pd.DB.AutoMigrate(&flavorEntity{}, &imageEntity{}, &reportEntity{})
	pd.DB.Table("image_flavors").
		AddForeignKey("image_id", "images(id)", "CASCADE", "CASCADE").
		AddForeignKey("flavor_id", "flavors(id)", "CASCADE", "CASCADE").
		AddUniqueIndex("image_flavor_index", "image_id", "flavor_id")
	return nil
}

func (pd PostgresDatabase) Driver() *gorm.DB {
	log.Trace("repository/postgres/postgres_database:Driver() Entering")
	defer log.Trace("repository/postgres/postgres_database:Driver() Leaving")
	return pd.DB
}

func (pd PostgresDatabase) FlavorRepository() repository.FlavorRepository {
	log.Trace("repository/postgres/postgres_database:FlavorRepository() Entering")
	defer log.Trace("repository/postgres/postgres_database:FlavorRepository() Leaving")
	return flavorRepo{db: pd.DB}
}

func (pd PostgresDatabase) ReportRepository() repository.ReportRepository {
	log.Trace("repository/postgres/postgres_database:ReportRepository() Entering")
	defer log.Trace("repository/postgres/postgres_database:ReportRepository() Leaving")
	return reportRepo{db: pd.DB}
}

func (pd PostgresDatabase) ImageRepository() repository.ImageRepository {
	log.Trace("repository/postgres/postgres_database:ImageRepository() Entering")
	defer log.Trace("repository/postgres/postgres_database:ImageRepository() Leaving")
	return imageRepo{db: pd.DB}
}
