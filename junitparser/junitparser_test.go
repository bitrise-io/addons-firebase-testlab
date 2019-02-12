package junitparser_test

import (
	"testing"

	"github.com/bitrise-io/addons-firebase-testlab/junitparser"
	"github.com/stretchr/testify/require"
)

func Test_Junitparser_ParseToJSON_validXML(t *testing.T) {
	xml := []byte(`
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
`)

	// Exp JSON prettyfied
	// {
	//   "name": "JUnitXmlReporter.constructor",
	//   "package": "",
	//   "properties": {
	//     "compiler.debug": "on",
	//     "java.vendor": "Sun Microsystems Inc.",
	//     "project.jdk.classpath": "jdk.classpath.1.6"
	//   },
	//   "tests": [
	//     {
	//       "name": "should default path to an empty string",
	//       "classname": "JUnitXmlReporter.constructor",
	//       "duration": 6000000,
	//       "Status": "failed",
	//       "Error": {
	//         "message": "test failure",
	//         "body": "Assertion failed"
	//       }
	//     },
	//     {
	//       "name": "should default consolidate to true",
	//       "classname": "JUnitXmlReporter.constructor",
	//       "duration": 0,
	//       "Status": "skipped",
	//       "Error": null
	//     },
	//     {
	//       "name": "should default useDotNotation to true",
	//       "classname": "JUnitXmlReporter.constructor",
	//       "duration": 0,
	//       "Status": "passed",
	//       "Error": null
	//     }
	//   ],
	//   "totals": {
	//     "tests": 3,
	//     "passed": 1,
	//     "skipped": 1,
	//     "failed": 1,
	//     "error": 0,
	//     "duration": 6000000
	//   }
	// }

	expJSON := []byte(`{"name":"JUnitXmlReporter.constructor","package":"","properties":{"compiler.debug":"on","java.vendor":"Sun Microsystems Inc.","project.jdk.classpath":"jdk.classpath.1.6"},"tests":[{"name":"should default path to an empty string","classname":"JUnitXmlReporter.constructor","duration":6000000,"Status":"failed","Error":{"message":"test failure","body":"Assertion failed"}},{"name":"should default consolidate to true","classname":"JUnitXmlReporter.constructor","duration":0,"Status":"skipped","Error":null},{"name":"should default useDotNotation to true","classname":"JUnitXmlReporter.constructor","duration":0,"Status":"passed","Error":null}],"totals":{"tests":3,"passed":1,"skipped":1,"failed":1,"error":0,"duration":6000000}}`)
	p := junitparser.Parser{}
	json, err := p.ParseToJSON(xml)
	require.NoError(t, err)
	require.Equal(t, string(expJSON), string(json))
}

func Test_Junitparser_ParseToJSON_emptyXML(t *testing.T) {
	xml := []byte(`
    <?xml version="1.0" encoding="UTF-8"?>
    <testsuites>
    </testsuites>
`)

	p := junitparser.Parser{}
	json, err := p.ParseToJSON(xml)
	require.Nil(t, json)
	require.EqualError(t, err, "The test report is empty")

}

func Test_Junitparser_ParseToJSON_invalidXML(t *testing.T) {
	xml := []byte(`
	<xml version="1.0" encoding="UTF-8"?>
	<testsuites>
	</testsuites>
	`)

	p := junitparser.Parser{}
	json, err := p.ParseToJSON(xml)
	require.Nil(t, json)
	require.EqualError(t, err, "Parsing of test report failed: XML syntax error on line 2: expected attribute name in element")
}
