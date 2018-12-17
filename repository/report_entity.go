package repository

import (
	"encoding/json"
	"errors"
	"time"

	"intel/isecl/workload-service/model"

	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
)

type reportEntity struct {
	// alias gorm.Model
	model.Report
	
	CreatedAt time.Time
	ExpiresOn time.Time
	
	*model.VMTrustReport `gorm:"-"`
}

func (re *reportEntity) BeforeCreate(scope *gorm.Scope) error {
	
	id, err := uuid.NewV4()
	if err != nil {
		return errors.New("Unable to create uuid. ")
	}
	if err := scope.SetColumn("id", id.String()); err != nil {
		return err
	}

	if re.VMTrustReport == nil {
		return errors.New("Content must not be null")
	}
	// none of the below can be nil, as they are not pointers
	vmID := re.VMTrustReport.Manifest.VmInfo.VmID
	if len(vmID) == 0 {
		return errors.New("Instance uuid cannot be empty")
	}
	jsonData, err := json.Marshal(re.VMTrustReport)
	if err != nil {
		return err
	}
	if err := scope.SetColumn("vm_id", vmID); err != nil {
		return err
	}
	if err := scope.SetColumn("trust_report", jsonData); err != nil {
		return err
	}
	if err := scope.SetColumn("saml", ""); err != nil {
		return err
	}
	return nil
}

func (re *reportEntity) AfterFind(scope *gorm.Scope) error {
	// unmarshal Content into VMTrustReport
	var vtr model.VMTrustReport
	err := json.Unmarshal(re.TrustReport.RawMessage, &vtr)
	if err != nil {
		return err
	}
	re.VMTrustReport = &vtr
	return nil
}
