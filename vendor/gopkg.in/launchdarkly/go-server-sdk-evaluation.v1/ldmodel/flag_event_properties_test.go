package ldmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldreason"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldtime"
	ldevents "gopkg.in/launchdarkly/go-sdk-events.v1"
)

func asFlagEventProperties(f FeatureFlag) ldevents.FlagEventProperties {
	return &f
}

func TestFlagEventPropertiesBasicProperties(t *testing.T) {
	flag := FeatureFlag{Key: "key", Version: 2}
	assert.Equal(t, "key", asFlagEventProperties(flag).GetKey())
	assert.Equal(t, 2, asFlagEventProperties(flag).GetVersion())
}

func TestIsFullEventTrackingEnabled(t *testing.T) {
	flag1 := FeatureFlag{Key: "key"}
	assert.False(t, asFlagEventProperties(flag1).IsFullEventTrackingEnabled())

	flag2 := FeatureFlag{Key: "key", TrackEvents: true}
	assert.True(t, asFlagEventProperties(flag2).IsFullEventTrackingEnabled())
}

func TestGetDebugEventsUntilDate(t *testing.T) {
	flag1 := FeatureFlag{Key: "key"}
	assert.Equal(t, ldtime.UnixMillisecondTime(0), asFlagEventProperties(flag1).GetDebugEventsUntilDate())

	date := ldtime.UnixMillisecondTime(100000)
	flag2 := FeatureFlag{Key: "key", DebugEventsUntilDate: date}
	assert.Equal(t, date, asFlagEventProperties(flag2).GetDebugEventsUntilDate())
}

func TestIsExperimentDefaultsToFalse(t *testing.T) {
	flag := FeatureFlag{Key: "key"}
	assert.False(t, asFlagEventProperties(flag).IsExperimentationEnabled(ldreason.NewEvalReasonOff()))
}

func TestIsExperimentReturnsFalseForFallthroughIfTrackEventsFallthroughIsFalse(t *testing.T) {
	flag := FeatureFlag{Key: "key"}
	assert.False(t, asFlagEventProperties(flag).IsExperimentationEnabled(ldreason.NewEvalReasonFallthrough()))
}

func TestIsExperimentReturnsTrueForFallthroughIfTrackEventsFallthroughIsTrue(t *testing.T) {
	flag := FeatureFlag{Key: "key", TrackEventsFallthrough: true}
	assert.True(t, asFlagEventProperties(flag).IsExperimentationEnabled(ldreason.NewEvalReasonFallthrough()))
}

func TestIsExperimentReturnsFalseForRuleMatchIfTrackEventsIsFalseForThatRule(t *testing.T) {
	flag := FeatureFlag{
		Key: "key",
		Rules: []FlagRule{
			{ID: "rule0", TrackEvents: true},
			{ID: "rule1", TrackEvents: false},
		},
	}
	reason := ldreason.NewEvalReasonRuleMatch(1, "rule1")
	assert.False(t, asFlagEventProperties(flag).IsExperimentationEnabled(reason))
}

func TestIsExperimentReturnsTrueForRuleMatchIfTrackEventsIsTrueForThatRule(t *testing.T) {
	flag := FeatureFlag{
		Key: "key",
		Rules: []FlagRule{
			{ID: "rule0", TrackEvents: true},
			{ID: "rule1", TrackEvents: false},
		},
	}
	reason := ldreason.NewEvalReasonRuleMatch(0, "rule0")
	assert.True(t, asFlagEventProperties(flag).IsExperimentationEnabled(reason))
}

func TestIsExperimentReturnsFalseForRuleMatchIfRuleIndexIsNegative(t *testing.T) {
	flag := FeatureFlag{
		Key: "key",
		Rules: []FlagRule{
			{ID: "rule0", TrackEvents: true},
			{ID: "rule1", TrackEvents: false},
		},
	}
	reason := ldreason.NewEvalReasonRuleMatch(-1, "rule1")
	assert.False(t, asFlagEventProperties(flag).IsExperimentationEnabled(reason))
}

func TestIsExperimentReturnsFalseForRuleMatchIfRuleIndexIsTooHigh(t *testing.T) {
	flag := FeatureFlag{
		Key: "key",
		Rules: []FlagRule{
			{ID: "rule0", TrackEvents: true},
			{ID: "rule1", TrackEvents: false},
		},
	}
	reason := ldreason.NewEvalReasonRuleMatch(2, "rule1")
	assert.False(t, asFlagEventProperties(flag).IsExperimentationEnabled(reason))
}
