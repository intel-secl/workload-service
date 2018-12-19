package model

import (
	 "github.com/jinzhu/gorm/dialects/postgres"
	 "intel/isecl/lib/common/pkg/vm"
	)

type Report struct{
    VmID      string `gorm:"type:uuid;not null" json:"vm_id"`
	ID        string `gorm:"type:uuid;primary_key;" json:"id"`
	Saml               string `json:"saml"`       
	TrustReport             postgres.Jsonb `gorm:"type:jsonb;not null"`
}

// Cannot use the VMReportTrust from verifier library because there is conflict in json unmarshalling as it is trying to type verifier.rule to Result.rule
// So explicitly creating a type for VMTrustReport  
type VMTrustReport struct {
	Manifest   vm.Manifest `json:"vm_manifest"`
	PolicyName string      `json:"policy_name"`
	Results    []Result    `json:"results"`
	Trusted    bool        `json:"trusted"`
}

// Result is a struct that indicates the evaluation conclusion of applying a rule against a manifest.
// The FlavorID from which the rule derived from is included as well.
type Result struct {
	// Rule is an interface, and can be any concrete interface. You will need to apply a type assertion based on what it is if you need to access it's fields.
	Rule     Rule    `json:"rule"`
	FlavorID string  `json:"flavor_id"`
	Faults   []Fault `json:"faults,omitempty"`
	Trusted  bool    `json:"trusted"`
}

// Fault defines failure events when applying a Rule
type Fault struct {
	Description string `json:"description"`
	Cause       error  `json:"cause"`
}

// EncryptionMatches is a rule that enforced VM image encryption policy from
type Rule struct {
	RuleName string             `json:"rule_name"`
	Markers  []string           `json:"markers"`
	Expected ExpectedEncryption `json:"expected"`
}

// ExpectedEncryption is a data template that defines the json tag name of the encryption requirement, and the expected boolean value
type ExpectedEncryption struct {
	Name  string `json:"name"`
	Value bool   `json:"Value"`
}
