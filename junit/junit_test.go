package junit_test

import (
	"testing"

	"github.com/bitrise-io/addons-firebase-testlab/junit"
	junitmodels "github.com/bitrise-io/go-junit"
	"github.com/stretchr/testify/require"
)

func Test_Junitparser_Parse(t *testing.T) {
	testCases := []struct {
		name    string
		xml     string
		expResp []junitmodels.Suite
		expErr  string
	}{
		{
			name: "when XML is valid and has test suites",
			xml: `
			<?xml version="1.0" encoding="UTF-8"?>
			<testsuites>
				<testsuite name="JUnitXmlReporter.constructor" errors="0" skipped="1" tests="3" failures="1" time="0.006" timestamp="2013-05-24T10:23:58">
					<properties>
						<property name="java.vendor" value="Sun Microsystems Inc." />
						<property name="compiler.debug" value="on" />
						<property name="project.jdk.classpath" value="jdk.classpath.1.6" />
					</properties>
					<testcase classname="JUnitXmlReporter.constructor" name="should default path to an empty string" time="0.006">
						<failure message="test failure">Assertion failed</failure>
					</testcase>
					<testcase classname="JUnitXmlReporter.constructor" name="should default consolidate to true" time="0">
						<skipped />
					</testcase>
					<testcase classname="JUnitXmlReporter.constructor" name="should default useDotNotation to true" time="0" />
				</testsuite>
			</testsuites>
			`,
			expResp: []junitmodels.Suite{
				junitmodels.Suite{
					Name:    "JUnitXmlReporter.constructor",
					Package: "",
					Properties: map[string]string{
						"java.vendor":           "Sun Microsystems Inc.",
						"compiler.debug":        "on",
						"project.jdk.classpath": "jdk.classpath.1.6"},
					SystemOut: "",
					SystemErr: "",
					Totals: junitmodels.Totals{
						Tests:    3,
						Passed:   1,
						Skipped:  1,
						Failed:   1,
						Error:    0,
						Duration: 6000000,
					},
					Tests: []junitmodels.Test{
						junitmodels.Test{
							Name:      "should default path to an empty string",
							Classname: "JUnitXmlReporter.constructor",
							Duration:  6000000, Status: "failed",
							Error: junitmodels.Error{
								Message: "test failure",
								Type:    "",
								Body:    "Assertion failed",
							},
							Properties: map[string]string{
								"classname": "JUnitXmlReporter.constructor",
								"name":      "should default path to an empty string",
								"time":      "0.006",
							},
						},
						junitmodels.Test{
							Name:      "should default consolidate to true",
							Classname: "JUnitXmlReporter.constructor",
							Duration:  0,
							Status:    "skipped",
							Error:     junitmodels.Error{},
							Properties: map[string]string{
								"classname": "JUnitXmlReporter.constructor",
								"name":      "should default consolidate to true",
								"time":      "0",
							},
						},
						junitmodels.Test{
							Name:      "should default useDotNotation to true",
							Classname: "JUnitXmlReporter.constructor",
							Duration:  0,
							Status:    "passed",
							Error:     junitmodels.Error{},
							Properties: map[string]string{
								"classname": "JUnitXmlReporter.constructor",
								"name":      "should default useDotNotation to true",
								"time":      "0",
							},
						},
					},
				},
			},
			expErr: "",
		},
		{
			name: "when test suites are empty",
			xml: `
			<?xml version="1.0" encoding="UTF-8"?>
			<testsuites>
			</testsuites>
			`,
			expResp: []junitmodels.Suite{},
			expErr:  "",
		},
		{
			name:   "when XML is invalid",
			xml:    "<xml?>",
			expErr: "Parsing of test report failed: XML syntax error on line 1: expected attribute name in element",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := junit.Client{}
			got, err := c.Parse([]byte(tc.xml))
			if len(tc.expErr) > 0 {
				require.EqualError(t, err, tc.expErr)

			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expResp, got)
			}
		})
	}
}
