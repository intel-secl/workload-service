/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package postgres

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"intel/isecl/workload-service/repository"
	"strings"
	"time"
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

func (pd *PostgresDatabase) Close() {
	log.Trace("repository/postgres/postgres_database:Close() Entering")
	defer log.Trace("repository/postgres/postgres_database:Close() Leaving")
	if pd.DB != nil {
		pd.DB.Close()
	}
}

func Open(host string, port int, dbname, user, password, sslMode, sslCert string) (*PostgresDatabase, error) {

	log.Trace("repository/postgres/postgres_database:Open() Entering")
	defer log.Trace("repository/postgres/postgres_database:Open() Leaving")

	sslMode = strings.TrimSpace(strings.ToLower(sslMode))
	if sslMode != "disable" && sslMode != "allow" && sslMode != "prefer" && sslMode != "verify-ca" && sslMode != "verify-full" {
		sslMode = "require"
	}

	var sslCertParams string
	if sslMode == "verify-ca" || sslMode == "verify-full" {
		sslCertParams = " sslrootcert=" + sslCert
	}

	var db *gorm.DB
	var dbErr error
	const numAttempts = 4
	for i := 0; i < numAttempts; i = i + 1 {
		const retryTime = 1
		db, dbErr = gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s%s",
			host, port, user, dbname, password, sslMode, sslCertParams))
		if dbErr != nil {
			log.WithError(dbErr).Infof("Failed to connect to DB, retrying attempt %d/%d", i, numAttempts)
		} else {
			break
		}
		time.Sleep(retryTime * time.Second)
	}
	if dbErr != nil {
		return nil, errors.Wrapf(dbErr, "Failed to connect to db after %d attempts", numAttempts)
	}
	return &PostgresDatabase{DB: db}, nil
}