package postgres

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"intel/isecl/workload-service/model"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/satori/go.uuid"
)

type reportEntity struct {
	ID        string `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time `sql:"type:timestamp"`
	ExpiresOn time.Time `sql:"type:timestamp"`
	// normalize VMID
	VMID        string `gorm:"type:uuid;not null"`
	Saml        string
	TrustReport postgres.Jsonb `gorm:"type:jsonb;not null"`
	SignedData  postgres.Jsonb `gorm:"type:jsonb;not null"`
}

func (re reportEntity) TableName() string {
	return "reports"
}

func (re *reportEntity) BeforeCreate(scope *gorm.Scope) error {

	id, err := uuid.NewV4()
	if err != nil {
		return errors.New("Unable to create uuid.")
	}
	if err := scope.SetColumn("id", id.String()); err != nil {
		return err
	}

	if !json.Valid(re.TrustReport.RawMessage) {
		return errors.New("trust report json content is not valid")
	}

	if !json.Valid(re.SignedData.RawMessage) {
		return errors.New("signed data json content is not valid")
	}

	return nil
}

func (re *reportEntity) AfterFind(scope *gorm.Scope) error {
	_, err := re.unmarshal()
	if err != nil {
		return errors.New("JSON Content does not match TrustReport schema")
	}
	return nil
}

func (re *reportEntity) unmarshal() (*model.Report, error) {
	var report model.Report

	if err := json.Unmarshal(re.TrustReport.RawMessage, &report.ImageTrustReport); err != nil {
		fmt.Println(string(re.TrustReport.RawMessage))
		fmt.Println(err)
		return nil, err
	}
	if err := json.Unmarshal(re.SignedData.RawMessage, &report.SignedData); err != nil {
		fmt.Println(string(re.SignedData.RawMessage))
		fmt.Println(err)
		return nil, err
	}
	report.ID = re.ID
	return &report, nil
}

func (re *reportEntity) Report() model.Report {
	r, _ := re.unmarshal()
	return *r
}
