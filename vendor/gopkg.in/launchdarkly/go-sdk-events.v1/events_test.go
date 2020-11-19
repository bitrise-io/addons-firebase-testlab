package ldevents

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldreason"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldtime"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

var defaultEventFactory = NewEventFactory(false, nil)

var noReason = ldreason.EvaluationReason{}

// Stub implementation of FlagEventProperties
type flagEventPropertiesImpl struct {
	Key                  string
	Version              int
	TrackEvents          bool
	DebugEventsUntilDate ldtime.UnixMillisecondTime
	IsExperiment         bool
}

func (f flagEventPropertiesImpl) GetKey() string                   { return f.Key }
func (f flagEventPropertiesImpl) GetVersion() int                  { return f.Version }
func (f flagEventPropertiesImpl) IsFullEventTrackingEnabled() bool { return f.TrackEvents }
func (f flagEventPropertiesImpl) GetDebugEventsUntilDate() ldtime.UnixMillisecondTime {
	return f.DebugEventsUntilDate
}
func (f flagEventPropertiesImpl) IsExperimentationEnabled(reason ldreason.EvaluationReason) bool {
	return f.IsExperiment
}

func TestEventFactory(t *testing.T) {
	fakeTime := ldtime.UnixMillisecondTime(100000)
	timeFn := func() ldtime.UnixMillisecondTime { return fakeTime }
	withoutReasons := NewEventFactory(false, timeFn)
	withReasons := NewEventFactory(true, timeFn)
	user := User(lduser.NewUser("u"))

	t.Run("NewSuccessfulEvalEvent", func(t *testing.T) {
		flag := flagEventPropertiesImpl{Key: "flagkey", Version: 100}

		expected := FeatureRequestEvent{
			BaseEvent: BaseEvent{
				CreationDate: fakeTime,
				User:         user,
			},
			Key:       flag.Key,
			Version:   ldvalue.NewOptionalInt(flag.Version),
			Variation: ldvalue.NewOptionalInt(1),
			Value:     ldvalue.String("value"),
			Default:   ldvalue.String("default"),
			Reason:    ldreason.NewEvalReasonFallthrough(),
			PrereqOf:  ldvalue.NewOptionalString("pre"),
		}

		event1 := withoutReasons.NewEvalEvent(flag, user,
			ldreason.NewEvaluationDetail(expected.Value, expected.Variation.IntValue(), expected.Reason),
			expected.Default, "pre")
		assert.Equal(t, ldreason.EvaluationReason{}, event1.Reason)
		event1.Reason = expected.Reason
		assert.Equal(t, expected, event1)

		event2 := withReasons.NewEvalEvent(flag, user,
			ldreason.NewEvaluationDetail(expected.Value, expected.Variation.IntValue(), expected.Reason),
			expected.Default, "pre")
		assert.Equal(t, expected, event2)
	})

	t.Run("NewSuccessfulEvalEvent with tracking/debugging", func(t *testing.T) {
		flag := flagEventPropertiesImpl{Key: "flagkey", Version: 100}

		expected := FeatureRequestEvent{
			BaseEvent: BaseEvent{
				CreationDate: fakeTime,
				User:         user,
			},
			Key:       flag.Key,
			Version:   ldvalue.NewOptionalInt(flag.Version),
			Variation: ldvalue.NewOptionalInt(1),
			Value:     ldvalue.String("value"),
			Default:   ldvalue.String("default"),
		}

		flag1 := flag
		flag1.TrackEvents = true
		expected1 := expected
		expected1.TrackEvents = true
		event1 := withoutReasons.NewEvalEvent(flag1, user,
			ldreason.NewEvaluationDetail(expected.Value, expected.Variation.IntValue(), ldreason.NewEvalReasonFallthrough()),
			expected.Default, "")
		assert.Equal(t, expected1, event1)

		flag2 := flag
		flag2.DebugEventsUntilDate = ldtime.UnixMillisecondTime(200000)
		expected2 := expected
		expected2.DebugEventsUntilDate = flag2.DebugEventsUntilDate
		event2 := withoutReasons.NewEvalEvent(flag2, user,
			ldreason.NewEvaluationDetail(expected.Value, expected.Variation.IntValue(), ldreason.NewEvalReasonFallthrough()),
			expected.Default, "")
		assert.Equal(t, expected2, event2)
	})

	t.Run("NewSuccessfulEvalEvent with experimentation", func(t *testing.T) {
		flag := flagEventPropertiesImpl{Key: "flagkey", Version: 100, IsExperiment: true}

		expected := FeatureRequestEvent{
			BaseEvent: BaseEvent{
				CreationDate: fakeTime,
				User:         user,
			},
			Key:         flag.Key,
			Version:     ldvalue.NewOptionalInt(flag.Version),
			Variation:   ldvalue.NewOptionalInt(1),
			Value:       ldvalue.String("value"),
			Default:     ldvalue.String("default"),
			Reason:      ldreason.NewEvalReasonFallthrough(),
			TrackEvents: true,
		}

		event := withoutReasons.NewEvalEvent(flag, user,
			ldreason.NewEvaluationDetail(expected.Value, expected.Variation.IntValue(), ldreason.NewEvalReasonFallthrough()),
			expected.Default, "")
		assert.Equal(t, expected, event)
	})

	t.Run("NewUnknownFlagEvent", func(t *testing.T) {
		expected := FeatureRequestEvent{
			BaseEvent: BaseEvent{
				CreationDate: fakeTime,
				User:         user,
			},
			Key:     "unknown-key",
			Value:   ldvalue.String("default"),
			Default: ldvalue.String("default"),
			Reason:  ldreason.NewEvalReasonFallthrough(),
		}

		event1 := withoutReasons.NewUnknownFlagEvent(expected.Key, user, expected.Default, expected.Reason)
		assert.Equal(t, ldreason.EvaluationReason{}, event1.Reason)
		event1.Reason = expected.Reason
		assert.Equal(t, expected, event1)

		event2 := withReasons.NewUnknownFlagEvent(expected.Key, user, expected.Default, expected.Reason)
		assert.Equal(t, expected, event2)
	})

	t.Run("NewCustomEvent", func(t *testing.T) {
		expected := CustomEvent{
			BaseEvent: BaseEvent{
				CreationDate: fakeTime,
				User:         user,
			},
			Key:         "event-key",
			Data:        ldvalue.String("data"),
			HasMetric:   true,
			MetricValue: 2,
		}

		event := withoutReasons.NewCustomEvent(expected.Key, user, expected.Data, true, expected.MetricValue)
		assert.Equal(t, expected, event)
	})

	t.Run("NewIdentifyEvent", func(t *testing.T) {
		expected := IdentifyEvent{
			BaseEvent: BaseEvent{
				CreationDate: fakeTime,
				User:         user,
			},
		}

		event := withoutReasons.NewIdentifyEvent(user)
		assert.Equal(t, expected, event)
	})

	t.Run("indexEvent (not from factory)", func(t *testing.T) {
		ie := indexEvent{
			BaseEvent: BaseEvent{
				CreationDate: fakeTime,
				User:         user,
			},
		}
		assert.Equal(t, ie.BaseEvent, ie.GetBase())
	})
}
