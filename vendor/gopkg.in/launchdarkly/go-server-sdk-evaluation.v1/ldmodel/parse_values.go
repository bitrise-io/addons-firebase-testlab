package ldmodel

import (
	"regexp"
	"time"

	"github.com/launchdarkly/go-semver"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

func parseDateTime(value ldvalue.Value) (time.Time, bool) {
	switch value.Type() {
	case ldvalue.StringType:
		return parseRFC3339TimeUTC(value.StringValue())
	case ldvalue.NumberType:
		return unixMillisToUtcTime(value.Float64Value()), true
	}
	return time.Time{}, false
}

func unixMillisToUtcTime(unixMillis float64) time.Time {
	return time.Unix(0, int64(unixMillis)*int64(time.Millisecond)).UTC()
}

func parseRegexp(value ldvalue.Value) (*regexp.Regexp, bool) {
	if value.Type() == ldvalue.StringType {
		if r, err := regexp.Compile(value.StringValue()); err == nil {
			return r, true
		}
	}
	return nil, false
}

func parseSemVer(value ldvalue.Value) (semver.Version, bool) {
	if value.Type() == ldvalue.StringType {
		versionStr := value.StringValue()
		if sv, err := semver.ParseAs(versionStr, semver.ParseModeAllowMissingMinorAndPatch); err == nil {
			return sv, true
		}
	}
	return semver.Version{}, false
}
