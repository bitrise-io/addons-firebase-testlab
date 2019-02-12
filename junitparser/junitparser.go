package junitparser

import (
	"encoding/json"

	junit "github.com/joshdk/go-junit"
	"github.com/pkg/errors"
)

// Parser ...
type Parser struct{}

// ParseToJSON ...
func (p Parser) ParseToJSON(xml []byte) ([]byte, error) {
	suites, err := junit.Ingest(xml)
	if err != nil {
		return nil, errors.Wrap(err, "Parsing of test report failed")
	}

	if (len(suites)) == 0 {
		return nil, errors.New("The test report is empty")
	}

	// TODO: Add support for multiple test suites in one XML file.
	suite := suites[0]

	jsonSuites, err := json.Marshal(suite)
	if err != nil {
		return nil, errors.Wrap(err, "JSON Marshaling of test report failed")
	}

	return jsonSuites, nil
}
