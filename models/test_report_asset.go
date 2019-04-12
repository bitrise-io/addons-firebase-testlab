package models

import (
	"fmt"
	"time"

	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
	uuid "github.com/satori/go.uuid"
)

// TestReportAsset ...
type TestReportAsset struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Filename     string     `json:"filename" db:"filename"`
	Filesize     int        `json:"filesize" db:"filesize"`
	Uploaded     bool       `json:"uploaded" db:"uploaded"`
	TestReport   TestReport `belongs_to:"test_report" json:"-" db:"-"`
	TestReportID uuid.UUID  `json:"test_report_id" db:"test_report_id"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Validate ...
func (t *TestReportAsset) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: t.Filename, Name: "Filename"},
		&validators.IntIsGreaterThan{Field: t.Filesize, Compared: 0, Name: "Filesize"},
	), nil
}

// PathInBucket ...
func (t *TestReportAsset) PathInBucket() string {
	return fmt.Sprintf("builds/%s/test_reports/%s/assets/%s", t.TestReport.BuildSlug, t.ID, t.Filename)
}
