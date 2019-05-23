package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/markbates/pop/nulls"
)

// Build ...
type Build struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	AppSlug             string     `json:"app_slug" db:"app_slug"`
	BuildSlug           string     `json:"build_slug" db:"build_slug"`
	BuildSessionEnabled bool       `json:"build_session_enabled" db:"build_session_enabled"`
	TestStartTime       nulls.Time `json:"test_start_time" db:"test_start_time"`
	TestEndTime         nulls.Time `json:"test_end_time" db:"test_end_time"`
	TestMatrixID        string     `json:"test_matrix_id" db:"test_matrix_id"`
	TestHistoryID       string     `json:"test_history_id" db:"test_history_id"`
	TestExecutionID     string     `json:"test_execution_id" db:"test_execution_id"`
	LastRequest         nulls.Time `json:"last_request" db:"last_request"`
}

// String is not required by pop and may be deleted
func (b Build) String() string {
	jb, err := json.Marshal(b)
	if err != nil {
		return ""
	}
	return string(jb)
}

// Builds is not required by pop and may be deleted
type Builds []Build

// String is not required by pop and may be deleted
func (b Builds) String() string {
	jb, err := json.Marshal(b)
	if err != nil {
		return ""
	}
	return string(jb)
}

// Validate gets run everytime you call a "pop.Validate" method.
// This method is not required and may be deleted.
func (b *Build) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: b.BuildSlug, Name: "BuildSlug"},
		&validators.StringIsPresent{Field: b.AppSlug, Name: "AppSlug"},
	), nil
}

// ValidateSave gets run everytime you call "pop.ValidateSave" method.
// This method is not required and may be deleted.
func (b *Build) ValidateSave(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run everytime you call "pop.ValidateUpdate" method.
// This method is not required and may be deleted.
func (b *Build) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
