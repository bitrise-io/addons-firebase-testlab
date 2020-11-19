package ldevents

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldtime"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

var user = EventUser{lduser.NewUser("key"), nil}
var undefInt = ldvalue.OptionalInt{}

func makeEvalEvent(creationDate ldtime.UnixMillisecondTime, flagKey string,
	flagVersion ldvalue.OptionalInt, variation ldvalue.OptionalInt, value, defaultValue string) FeatureRequestEvent {
	return FeatureRequestEvent{
		BaseEvent: BaseEvent{CreationDate: creationDate, User: user},
		Key:       flagKey,
		Version:   flagVersion,
		Variation: variation,
		Value:     ldvalue.String(value),
		Default:   ldvalue.String(defaultValue),
	}
}

func TestSummarizeEventSetsStartAndEndDates(t *testing.T) {
	es := newEventSummarizer()
	flagKey := "key"
	event1 := makeEvalEvent(2000, flagKey, ldvalue.NewOptionalInt(1), ldvalue.NewOptionalInt(0), "", "")
	event2 := makeEvalEvent(1000, flagKey, ldvalue.NewOptionalInt(1), ldvalue.NewOptionalInt(0), "", "")
	event3 := makeEvalEvent(1500, flagKey, ldvalue.NewOptionalInt(1), ldvalue.NewOptionalInt(0), "", "")
	es.summarizeEvent(event1)
	es.summarizeEvent(event2)
	es.summarizeEvent(event3)
	data := es.snapshot()

	assert.Equal(t, ldtime.UnixMillisecondTime(1000), data.startDate)
	assert.Equal(t, ldtime.UnixMillisecondTime(2000), data.endDate)
}

func TestSummarizeEventIncrementsCounters(t *testing.T) {
	es := newEventSummarizer()
	flagKey1 := "key1"
	flagKey2 := "key2"
	flagVersion1 := ldvalue.NewOptionalInt(11)
	flagVersion2 := ldvalue.NewOptionalInt(22)

	unknownFlagKey := "badkey"
	variation1 := ldvalue.NewOptionalInt(1)
	variation2 := ldvalue.NewOptionalInt(2)
	event1 := makeEvalEvent(0, flagKey1, flagVersion1, variation1, "value1", "default1")
	event2 := makeEvalEvent(0, flagKey1, flagVersion1, variation2, "value2", "default1")
	event3 := makeEvalEvent(0, flagKey2, flagVersion2, variation1, "value99", "default2")
	event4 := makeEvalEvent(0, flagKey1, flagVersion1, variation1, "value1", "default1")
	event5 := makeEvalEvent(0, unknownFlagKey, undefInt, undefInt, "default3", "default3")

	es.summarizeEvent(event1)
	es.summarizeEvent(event2)
	es.summarizeEvent(event3)
	es.summarizeEvent(event4)
	es.summarizeEvent(event5)
	data := es.snapshot()

	expectedCounters := map[counterKey]*counterValue{
		counterKey{flagKey1, variation1, flagVersion1}: &counterValue{2, ldvalue.String("value1"), ldvalue.String("default1")},
		counterKey{flagKey1, variation2, flagVersion1}: &counterValue{1, ldvalue.String("value2"), ldvalue.String("default1")},
		counterKey{flagKey2, variation1, flagVersion2}: &counterValue{1, ldvalue.String("value99"), ldvalue.String("default2")},
		counterKey{unknownFlagKey, undefInt, undefInt}: &counterValue{1, ldvalue.String("default3"), ldvalue.String("default3")},
	}
	assert.Equal(t, expectedCounters, data.counters)
}

func TestCounterForNilVariationIsDistinctFromOthers(t *testing.T) {
	es := newEventSummarizer()
	flagKey := "key1"
	flagVersion := ldvalue.NewOptionalInt(11)
	variation1 := ldvalue.NewOptionalInt(1)
	variation2 := ldvalue.NewOptionalInt(2)
	event1 := makeEvalEvent(0, flagKey, flagVersion, variation1, "value1", "default1")
	event2 := makeEvalEvent(0, flagKey, flagVersion, variation2, "value2", "default1")
	event3 := makeEvalEvent(0, flagKey, flagVersion, undefInt, "default1", "default1")
	es.summarizeEvent(event1)
	es.summarizeEvent(event2)
	es.summarizeEvent(event3)
	data := es.snapshot()

	expectedCounters := map[counterKey]*counterValue{
		counterKey{flagKey, variation1, flagVersion}: &counterValue{1, ldvalue.String("value1"), ldvalue.String("default1")},
		counterKey{flagKey, variation2, flagVersion}: &counterValue{1, ldvalue.String("value2"), ldvalue.String("default1")},
		counterKey{flagKey, undefInt, flagVersion}:   &counterValue{1, ldvalue.String("default1"), ldvalue.String("default1")},
	}
	assert.Equal(t, expectedCounters, data.counters)
}
