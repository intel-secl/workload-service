/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package postgres

import (
	"encoding/json"
	"fmt"
	"intel/isecl/workload-service/v3/model"
	"time"

	"github.com/pkg/errors"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/google/uuid"
)

type reportEntity struct {
	ID        string    `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time `sql:"type:timestamp"`
	ExpiresOn time.Time `sql:"type:timestamp"`
	// normalize InstanceID
	InstanceID  string `gorm:"type:uuid;not null"`
	Saml        string
	TrustReport postgres.Jsonb `gorm:"type:jsonb;not null"`
	SignedData  postgres.Jsonb `gorm:"type:jsonb;not null"`
}

func (re reportEntity) TableName() string {
	log.Trace("repository/postgres/report_entity:TableName() Entering")
	defer log.Trace("repository/postgres/report_entity:TableName() Leaving")
	return "reports"
}

func (re *reportEntity) BeforeCreate(scope *gorm.Scope) error {
	log.Trace("repository/postgres/report_entity:BeforeCreate() Entering")
	defer log.Trace("repository/postgres/report_entity:BeforeCreate() Leaving")

	id, err := uuid.NewV4()
	if err != nil {
		return errors.New("repository/postgres/report_entity:BeforeCreate() unable to create uuid")
	}
	if err := scope.SetColumn("id", id.String()); err != nil {
		return errors.New("repository/postgres/report_entity:BeforeCreate() unable to set column value")
	}

	if !json.Valid(re.TrustReport.RawMessage) {
		return errors.New("repository/postgres/report_entity:BeforeCreate() trust report json content is not valid")
	}

	if !json.Valid(re.SignedData.RawMessage) {
		return errors.New("repository/postgres/report_entity:BeforeCreate() signed data json content is not valid")
	}

	return nil
}

func (re *reportEntity) AfterFind(scope *gorm.Scope) error {
	log.Trace("repository/postgres/report_entity:AfterFind() Entering")
	defer log.Trace("repository/postgres/report_entity:AfterFind() Leaving")

	_, err := re.unmarshal()
	if err != nil {
		return errors.New("repository/postgres/report_entity:AfterFind() JSON Content does not match TrustReport schema")
	}
	return nil
}

func (re *reportEntity) unmarshal() (*model.Report, error) {
	log.Trace("repository/postgres/report_entity:unmarshal() Entering")
	defer log.Trace("repository/postgres/report_entity:unmarshal() Leaving")

	var report model.Report
	if err := json.Unmarshal(re.TrustReport.RawMessage, &report.InstanceTrustReport); err != nil {
		fmt.Println(string(re.TrustReport.RawMessage))
		fmt.Println(err)
		return nil, errors.Wrap(err, "repository/postgres/report_entity:unmarshal() Unable to unmarshal trust report")
	}
	if err := json.Unmarshal(re.SignedData.RawMessage, &report.SignedData); err != nil {
		fmt.Println(string(re.SignedData.RawMessage))
		fmt.Println(err)
		return nil, errors.Wrap(err, "repository/postgres/report_entity:unmarshal() Unable to unmarshal signed data")
	}
	report.ID = re.ID
	return &report, nil
}

func (re *reportEntity) Report() model.Report {
	log.Trace("repository/postgres/report_entity:Report() Entering")
	defer log.Trace("repository/postgres/report_entity:Report() Leaving")
	r, _ := re.unmarshal()
	return *r
}
