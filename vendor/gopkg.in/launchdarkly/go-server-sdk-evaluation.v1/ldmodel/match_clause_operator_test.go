package ldmodel

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

const dateStr1 = "2017-12-06T00:00:00.000-07:00"
const dateStr2 = "2017-12-06T00:01:01.000-07:00"
const dateMs1 = 10000000
const dateMs2 = 10000001
const invalidDate = "hey what's this?"

type opTestInfo struct {
	opName           Operator
	userValue        interface{}
	clauseValue      interface{}
	moreClauseValues []interface{}
	expected         bool
}

var operatorTests = []opTestInfo{
	// numeric operators
	{"in", int(99), int(99), nil, true},
	{"in", int(99), int(99), []interface{}{int(98), int(97), int(96)}, true},
	{"in", float64(99.0001), float64(99.0001), nil, true},
	{"in", float64(99.0001), float64(99.0001), []interface{}{float64(98), float64(97), float64(96)}, true},
	{"lessThan", int(1), float64(1.99999), nil, true},
	{"lessThan", float64(1.99999), int(1), nil, false},
	{"lessThan", int(1), uint(2), nil, true},
	{"lessThanOrEqual", int(1), float64(1), nil, true},
	{"greaterThan", int(2), float64(1.99999), nil, true},
	{"greaterThan", float64(1.99999), int(2), nil, false},
	{"greaterThan", int(2), uint(1), nil, true},
	{"greaterThanOrEqual", int(1), float64(1), nil, true},

	// string operators
	{"in", "x", "x", nil, true},
	{"in", "x", "x", []interface{}{"a", "b", "c"}, true},
	{"in", "x", "xyz", nil, false},
	{"startsWith", "xyz", "x", nil, true},
	{"startsWith", "x", "xyz", nil, false},
	{"endsWith", "xyz", "z", nil, true},
	{"endsWith", "z", "xyz", nil, false},
	{"contains", "xyz", "y", nil, true},
	{"contains", "y", "xyz", nil, false},

	// mixed strings and numbers
	{"in", "99", int(99), nil, false},
	{"in", int(99), "99", nil, false},
	{"contains", "99", int(99), nil, false},
	{"startsWith", "99", int(99), nil, false},
	{"endsWith", "99", int(99), nil, false},
	{"lessThanOrEqual", "99", int(99), nil, false},
	{"lessThanOrEqual", int(99), "99", nil, false},
	{"greaterThanOrEqual", "99", int(99), nil, false},
	{"greaterThanOrEqual", int(99), "99", nil, false},

	// equality of boolean values
	{"in", true, true, nil, true},
	{"in", false, false, nil, true},
	{"in", true, false, nil, false},
	{"in", false, true, nil, false},
	{"in", true, false, []interface{}{true}, true},

	// equality of JSON array values - note that the user value must be an array *of arrays*, because a single-level
	// array is interpreted as "any of these values"
	{"in", []interface{}{[]interface{}{"x"}}, []interface{}{"x"}, nil, true},
	{"in", []interface{}{[]interface{}{"x"}}, []interface{}{"x"}, []interface{}{[]interface{}{"a"}, []interface{}{"b"}}, true},

	// equality of JSON object values
	{"in", map[string]interface{}{"x": "1"}, map[string]interface{}{"x": "1"}, nil, true},
	{"in", map[string]interface{}{"x": "1"}, map[string]interface{}{"x": "1"},
		[]interface{}{map[string]interface{}{"a": "2"}, map[string]interface{}{"b": "3"}}, true},

	// regex
	{"matches", "hello world", "hello.*rld", nil, true},
	{"matches", "hello world", "hello.*orl", nil, true},
	{"matches", "hello world", "l+", nil, true},
	{"matches", "hello world", "(world|planet)", nil, true},
	{"matches", "hello world", "aloha", nil, false},
	{"matches", "hello world", "***bad regex", nil, false},

	// date operators
	{"before", dateStr1, dateStr2, nil, true},
	{"before", dateMs1, dateMs2, nil, true},
	{"before", dateStr2, dateStr1, nil, false},
	{"before", dateMs2, dateMs1, nil, false},
	{"before", dateStr1, dateStr1, nil, false},
	{"before", dateMs1, dateMs1, nil, false},
	{"before", nil, dateStr1, nil, false},
	{"before", dateStr1, invalidDate, nil, false},
	{"after", dateStr2, dateStr1, nil, true},
	{"after", dateMs2, dateMs1, nil, true},
	{"after", dateStr1, dateStr2, nil, false},
	{"after", dateMs1, dateMs2, nil, false},
	{"after", dateStr1, dateStr1, nil, false},
	{"after", dateMs1, dateMs1, nil, false},
	{"after", nil, dateStr1, nil, false},
	{"after", dateStr1, invalidDate, nil, false},

	// semver operators
	{"semVerEqual", "2.0.0", "2.0.0", nil, true},
	{"semVerEqual", "2.0", "2.0.0", nil, true},
	{"semVerEqual", "2-rc1", "2.0.0-rc1", nil, true},
	{"semVerEqual", "2+build2", "2.0.0+build2", nil, true},
	{"semVerEqual", "2.0.0", "2.0.1", nil, false},
	{"semVerLessThan", "2.0.0", "2.0.1", nil, true},
	{"semVerLessThan", "2.0", "2.0.1", nil, true},
	{"semVerLessThan", "2.0.1", "2.0.0", nil, false},
	{"semVerLessThan", "2.0.1", "2.0", nil, false},
	{"semVerLessThan", "2.0.1", "xbad%ver", nil, false},
	{"semVerLessThan", "2.0.0-rc", "2.0.0-rc.beta", nil, true},
	{"semVerGreaterThan", "2.0.1", "2.0", nil, true},
	{"semVerGreaterThan", "2.0.1", "2.0", nil, true},
	{"semVerGreaterThan", "2.0.0", "2.0.1", nil, false},
	{"semVerGreaterThan", "2.0", "2.0.1", nil, false},
	{"semVerGreaterThan", "2.0.1", "xbad%ver", nil, false},
	{"semVerGreaterThan", "2.0.0-rc.1", "2.0.0-rc.0", nil, true},

	// invalid operator
	{"whatever", "x", "x", nil, false},
}

func TestAllOperators(t *testing.T) {
	userAttr := "attr"
	for _, ti := range operatorTests {
		for _, withPreprocessing := range []bool{false, true} {
			t.Run(
				fmt.Sprintf("%v %s %v should be %v (preprocess: %t)", ti.userValue, ti.opName, ti.clauseValue, ti.expected, withPreprocessing),
				func(t *testing.T) {
					uValue := ldvalue.CopyArbitraryValue(ti.userValue)
					cValue := ldvalue.CopyArbitraryValue(ti.clauseValue)
					c := Clause{Attribute: lduser.UserAttribute(userAttr), Op: ti.opName}
					for _, v := range ti.moreClauseValues {
						c.Values = append(c.Values, ldvalue.CopyArbitraryValue(v))
					}
					c.Values = append(c.Values, cValue)
					if withPreprocessing {
						c.preprocessed = preprocessClause(c)
					}
					user := lduser.NewUserBuilder("key").Custom(userAttr, uValue).Build()
					isMatch := ClauseMatchesUser(&c, &user)
					assert.Equal(t, ti.expected, isMatch)
				},
			)
		}
	}
}

func TestParseDateZero(t *testing.T) {
	expectedTimeStamp := "1970-01-01T00:00:00Z"
	expected, _ := time.Parse(time.RFC3339Nano, expectedTimeStamp)
	testParseTime(t, expected, expected)
	testParseTime(t, 0, expected)
	testParseTime(t, 0.0, expected)
	testParseTime(t, expectedTimeStamp, expected)
}

func TestParseUtcTimestamp(t *testing.T) {
	expectedTimeStamp := "2016-04-16T22:57:31.684Z"
	expected, _ := time.Parse(time.RFC3339Nano, expectedTimeStamp)
	testParseTime(t, expected, expected)
	testParseTime(t, 1460847451684, expected)
	testParseTime(t, 1460847451684.0, expected)
	testParseTime(t, expectedTimeStamp, expected)
}

func TestParseTimezone(t *testing.T) {
	expectedTimeStamp := "2016-04-16T17:09:12.759-07:00"
	expected, _ := time.Parse(time.RFC3339Nano, expectedTimeStamp)
	testParseTime(t, expected, expected)
	testParseTime(t, 1460851752759, expected)
	testParseTime(t, 1460851752759.0, expected)
	testParseTime(t, expectedTimeStamp, expected)
}

func TestParseTimezoneNoMillis(t *testing.T) {
	expectedTimeStamp := "2016-04-16T17:09:12-07:00"
	expected, _ := time.Parse(time.RFC3339Nano, expectedTimeStamp)
	testParseTime(t, expected, expected)
	testParseTime(t, 1460851752000, expected)
	testParseTime(t, 1460851752000.0, expected)
	testParseTime(t, expectedTimeStamp, expected)
}

func TestParseTimestampBeforeEpoch(t *testing.T) {
	expectedTimeStamp := "1969-12-31T23:57:56.544-00:00"
	expected, _ := time.Parse(time.RFC3339Nano, expectedTimeStamp)
	testParseTime(t, expected, expected)
	testParseTime(t, -123456, expected)
	testParseTime(t, -123456.0, expected)
	testParseTime(t, expectedTimeStamp, expected)
}

func testParseTime(t *testing.T, input interface{}, expected time.Time) {
	expectedUTC := expected.UTC()
	value := ldvalue.CopyArbitraryValue(input)
	actual, ok := parseDateTime(value)
	if !ok {
		t.Errorf("failed to parse: %s", value)
		return
	}

	if !actual.Equal(expectedUTC) {
		t.Errorf("Got unexpected result: %+v Expected: %+v when parsing: %s", actual, expectedUTC, value)
	}
}
