package ldmodel

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseRFC3339TimeUTC(t *testing.T) {
	for _, test := range makeParseTimeTests() {
		t.Run(fmt.Sprintf("input string: [%s], valid: %t", test.s, test.valid), func(t *testing.T) {
			t1, ok := parseRFC3339TimeUTC(test.s)
			t2, err := time.Parse(time.RFC3339Nano, test.s)
			if test.valid {
				if assert.True(t, ok, "parseRFC3339TimeUTC should have accepted this string") &&
					assert.NoError(t, err, "time.Parse should have accepted this string") {
					assert.Equal(t, t2.UTC(), t1)
				}
			} else {
				if assert.False(t, ok, "parseRFC3339TimeUTC should have rejected this string") &&
					assert.Error(t, err, "time.Parse should have rejected this string") {
					assert.Equal(t, time.Time{}, t1)
				}
			}
		})
	}
}

type parseTimeTest struct {
	s     string
	valid bool
}

func makeParseTimeTests() []parseTimeTest {
	var ret []parseTimeTest

	addGood := func(s string) {
		ret = append(ret, parseTimeTest{s, true})
	}
	addBad := func(s string) {
		ret = append(ret, parseTimeTest{s, false})
	}

	for _, goodSuffix := range []string{
		"Z", ".123Z", ".123456789Z", "+08:00", "-08:00", "+07:30", "+99:00",
		// NONSTANDARD: time.Parse allows overly large hour offsets like "+99:00"
	} {
		addGood(fmt.Sprintf("2020-06-09T18:53:52%s", goodSuffix))
	}

	addGood("2020-06-09T1:53:52Z") // NONSTANDARD: time.Parse allows 1-digit hour

	for _, badYear := range []string{"202", "20202", "202x", ""} {
		addBad(fmt.Sprintf("%s-06-09T18:53:52Z", badYear))
	}
	for _, badMonth := range []string{"6", "006", "00", "13", "0x", ""} {
		addBad(fmt.Sprintf("2020-%s-09T18:53:52Z", badMonth))
	}
	for _, badDay := range []string{"9", "009", "00", "32", "3x", ""} {
		addBad(fmt.Sprintf("2020-06-%sT18:53:52Z", badDay))
	}
	for _, badHour := range []string{"018", "24", "1x", ""} {
		addBad(fmt.Sprintf("2020-06-09T%s:53:52Z", badHour))
	}
	for _, badMinute := range []string{"5", "053", "60", "5x", ""} {
		addBad(fmt.Sprintf("2020-06-09T18:%s:52Z", badMinute))
	}
	for _, badSecond := range []string{"5", "052", "61", "5x", ""} {
		addBad(fmt.Sprintf("2020-06-09T18:53:%sZ", badSecond))
	}
	for _, badFraction := range []string{".123456789123456789Z", ".500x", "."} {
		addBad(fmt.Sprintf("2020-06-09T18:53:52%sZ", badFraction))
	}
	for _, badTimeZone := range []string{
		"/07:00", "-0:00", "-007:00", "-2x:00", "-:00", "-07:0", "07:000", "07:60", "07:0x", "07:",
	} {
		addBad(fmt.Sprintf("2020-06-09T18:53:52%s", badTimeZone))
	}

	// RFC3339 doesn't support non-ASCII characters - just make sure our parser rejects them gracefully
	addBad("🤨2020-06-09T18:53:52Z")

	return ret
}

var parseTimeTests = []parseTimeTest{
	// bad fractional second
	{"2020-06-09T18:53:52.500xZ", false}, // non-numeric
	{"2020-06-09T18:53:52.Z", false},     // empty

	// bad time zone
	{"2020-06-09T18:53:52/07:00", false},  // invalid sign
	{"2020-06-09T18:53:52-0:00", false},   // hour too short
	{"2020-06-09T18:53:52-007:00", false}, // hour too long
	{"2020-06-09T18:53:52-24:00", false},  // hour too high
	{"2020-06-09T18:53:52-2x:00", false},  // hour non-numeric
	{"2020-06-09T18:53:52-:00", false},    // hour too short
}
