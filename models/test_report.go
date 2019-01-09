package models

import (
	"time"

	"github.com/satori/go.uuid"
)

// TestReport ...
type TestReport struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Filename  string    `json:"filename" db:"filename"`
	Filesize  int       `json:"filesize" db:"filesize"`
	Uploaded  bool      `json:"bool" db:"bool"`
	BuildSlug string    `json:"build_slug" db:"build_slug"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
