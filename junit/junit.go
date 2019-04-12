package junit

import (
	junitparser "github.com/joshdk/go-junit"
	"github.com/pkg/errors"
)

// Parser ...
type Parser interface {
	Parse(xml []byte) ([]junitparser.Suite, error)
}

// Client ...
type Client struct{}

// Parse ...
func (c *Client) Parse(xml []byte) ([]junitparser.Suite, error) {
	suites, err := junitparser.Ingest(xml)
	if err != nil {
		return []junitparser.Suite{}, errors.Wrap(err, "Parsing of test report failed")
	}

	return suites, nil
}
