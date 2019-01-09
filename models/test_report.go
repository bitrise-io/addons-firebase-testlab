package models

import (
	"fmt"
	"time"

	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
	"github.com/satori/go.uuid"
)

// TestReport ...
type TestReport struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Filename  string    `json:"filename" db:"filename"`
	Filesize  int       `json:"filesize" db:"filesize"`
	Uploaded  bool      `json:"uploaded" db:"uploaded"`
	BuildSlug string    `json:"build_slug" db:"build_slug"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}

// Validate ...
func (t *TestReport) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: t.Filename, Name: "Filename"},
		&validators.IntIsGreaterThan{Field: t.Filesize, Compared: 0, Name: "Filesize"},
	), nil
}

// PathInBucket ...
func (t *TestReport) PathInBucket() string {
	return fmt.Sprintf("builds/%s/test_reports/%s/%s", t.BuildSlug, t.ID, t.Filename)
}
