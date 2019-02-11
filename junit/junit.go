package junit

import (
	junitparser "github.com/joshdk/go-junit"
	"github.com/pkg/errors"
)

// Suite ...
type Suite struct {
	junitparser.Suite
}

// Parser ...
type Parser interface {
	Parse(xml []byte) ([]Suite, error)
}

// Client ...
type Client struct{}

// Parse ...
func (c *Client) Parse(xml []byte) ([]Suite, error) {
	rawSuites, err := junitparser.Ingest(xml)
	if err != nil {
		return []Suite{}, errors.Wrap(err, "Parsing of test report failed")
	}

	suites := []Suite{}
	for _, suite := range rawSuites {
		suites = append(suites, Suite{suite})
	}
	return suites, nil
}
