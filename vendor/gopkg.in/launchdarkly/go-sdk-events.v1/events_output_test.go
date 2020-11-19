package ldevents

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldreason"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldtime"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

const fakeTime = ldtime.UnixMillisecondTime(100000)

var (
	withoutReasons = NewEventFactory(false, fakeTimeFn)
	withReasons    = NewEventFactory(true, fakeTimeFn)
)

func fakeTimeFn() ldtime.UnixMillisecondTime { return fakeTime }

func TestEventOutputFullEvents(t *testing.T) {
	user := User(lduser.NewUser("u"))
	flag := flagEventPropertiesImpl{Key: "flagkey", Version: 100}

	defaultFormatter := eventOutputFormatter{config: epDefaultConfig}
	formatterWithInlineUsers := eventOutputFormatter{config: epDefaultConfig}
	formatterWithInlineUsers.config.InlineUsersInEvents = true

	t.Run("feature", func(t *testing.T) {
		event1 := withoutReasons.NewEvalEvent(flag, user, ldreason.NewEvaluationDetail(ldvalue.String("v"), 1, noReason),
			ldvalue.String("dv"), "")
		verifyEventOutput(t, defaultFormatter, event1,
			`{"kind":"feature","creationDate":100000,"key":"flagkey","version":100,"userKey":"u","variation":1,"value":"v","default":"dv"}`)
		verifyEventOutput(t, formatterWithInlineUsers, event1,
			`{"kind":"feature","creationDate":100000,"key":"flagkey","version":100,"user":{"key":"u"},"variation":1,"value":"v","default":"dv"}`)

		event1r := withReasons.NewEvalEvent(flag, user,
			ldreason.NewEvaluationDetail(ldvalue.String("v"), 1, ldreason.NewEvalReasonFallthrough()),
			ldvalue.String("dv"), "")
		verifyEventOutput(t, defaultFormatter, event1r,
			`{"kind":"feature","creationDate":100000,"key":"flagkey","version":100,"userKey":"u","variation":1,"value":"v","default":"dv","reason":{"kind":"FALLTHROUGH"}}`)

		event2 := withoutReasons.NewEvalEvent(flag, user, ldreason.EvaluationDetail{Value: ldvalue.String("v")},
			ldvalue.String("dv"), "")
		event2.Variation = ldvalue.OptionalInt{}
		verifyEventOutput(t, defaultFormatter, event2,
			`{"kind":"feature","creationDate":100000,"key":"flagkey","version":100,"userKey":"u","value":"v","default":"dv"}`)

		event3 := withoutReasons.NewEvalEvent(flag, user, ldreason.NewEvaluationDetail(ldvalue.String("v"), 1, noReason),
			ldvalue.String("dv"), "pre")
		verifyEventOutput(t, defaultFormatter, event3,
			`{"kind":"feature","creationDate":100000,"key":"flagkey","version":100,"userKey":"u","variation":1,"value":"v","default":"dv","prereqOf":"pre"}`)

		event4 := withoutReasons.NewUnknownFlagEvent("flagkey", user,
			ldvalue.String("dv"), ldreason.EvaluationReason{})
		verifyEventOutput(t, defaultFormatter, event4,
			`{"kind":"feature","creationDate":100000,"key":"flagkey","userKey":"u","value":"dv","default":"dv"}`)
	})

	t.Run("debug", func(t *testing.T) {
		event1 := withoutReasons.NewEvalEvent(flag, user, ldreason.NewEvaluationDetail(ldvalue.String("v"), 1, noReason),
			ldvalue.String("dv"), "")
		event1.Debug = true
		verifyEventOutput(t, defaultFormatter, event1,
			`{"kind":"debug","creationDate":100000,"key":"flagkey","version":100,"user":{"key":"u"},"variation":1,"value":"v","default":"dv"}`)
	})

	t.Run("identify", func(t *testing.T) {
		event := withoutReasons.NewIdentifyEvent(user)
		verifyEventOutput(t, defaultFormatter, event,
			`{"kind":"identify","creationDate":100000,"key":"u","user":{"key":"u"}}`)
	})

	t.Run("custom", func(t *testing.T) {
		event1 := withoutReasons.NewCustomEvent("eventkey", user, ldvalue.Null(), false, 0)
		verifyEventOutput(t, defaultFormatter, event1,
			`{"kind":"custom","creationDate":100000,"key":"eventkey","userKey":"u"}`)
		verifyEventOutput(t, formatterWithInlineUsers, event1,
			`{"kind":"custom","creationDate":100000,"key":"eventkey","user":{"key":"u"}}`)

		event2 := withoutReasons.NewCustomEvent("eventkey", user, ldvalue.String("d"), false, 0)
		verifyEventOutput(t, defaultFormatter, event2,
			`{"kind":"custom","creationDate":100000,"key":"eventkey","userKey":"u","data":"d"}`)

		event3 := withoutReasons.NewCustomEvent("eventkey", user, ldvalue.Null(), true, 2.5)
		verifyEventOutput(t, defaultFormatter, event3,
			`{"kind":"custom","creationDate":100000,"key":"eventkey","userKey":"u","metricValue":2.5}`)
	})

	t.Run("index", func(t *testing.T) {
		event := indexEvent{BaseEvent: BaseEvent{CreationDate: fakeTime, User: user}}
		verifyEventOutput(t, defaultFormatter, event,
			`{"kind":"index","creationDate":100000,"user":{"key":"u"}}`)
	})
}

func TestEventOutputSummaryEvents(t *testing.T) {
	user := User(lduser.NewUser("u"))
	flag1v1 := flagEventPropertiesImpl{Key: "flag1", Version: 100}
	flag1v2 := flagEventPropertiesImpl{Key: "flag1", Version: 200}
	flag1Default := ldvalue.String("default1")
	flag2 := flagEventPropertiesImpl{Key: "flag2", Version: 1}
	flag2Default := ldvalue.String("default2")

	defaultFormatter := eventOutputFormatter{config: epDefaultConfig}

	t.Run("summary - single flag, single counter", func(t *testing.T) {
		es1 := newEventSummarizer()
		event1 := withoutReasons.NewEvalEvent(flag1v1, user, ldreason.NewEvaluationDetail(ldvalue.String("v"), 1, noReason),
			ldvalue.String("dv"), "")
		es1.summarizeEvent(event1)
		verifySummaryEventOutput(t, defaultFormatter, es1.snapshot(),
			`{"kind":"summary","startDate":100000,"endDate":100000,"features":{"flag1":{"counters":[{"count":1,"value":"v","variation":1,"version":100}],"default":"dv"}}}`)

		es2 := newEventSummarizer()
		event2 := withoutReasons.NewEvalEvent(flag1v1, user, ldreason.EvaluationDetail{Value: ldvalue.String("dv")},
			ldvalue.String("dv"), "")
		event2.Variation = ldvalue.OptionalInt{}
		es2.summarizeEvent(event2)
		verifySummaryEventOutput(t, defaultFormatter, es2.snapshot(),
			`{"kind":"summary","startDate":100000,"endDate":100000,"features":{"flag1":{"counters":[{"count":1,"value":"dv","version":100}],"default":"dv"}}}`)

		es3 := newEventSummarizer()
		event3 := withoutReasons.NewUnknownFlagEvent("flagkey", user,
			ldvalue.String("dv"), ldreason.EvaluationReason{})
		es3.summarizeEvent(event3)
		verifySummaryEventOutput(t, defaultFormatter, es3.snapshot(),
			`{"kind":"summary","startDate":100000,"endDate":100000,"features":{"flagkey":{"counters":[{"count":1,"value":"dv","unknown":true}],"default":"dv"}}}`)
	})

	t.Run("summary - multiple counters", func(t *testing.T) {
		es := newEventSummarizer()
		es.summarizeEvent(withoutReasons.NewEvalEvent(flag1v1, user, ldreason.NewEvaluationDetail(ldvalue.String("a"), 1, noReason),
			flag1Default, ""))
		es.summarizeEvent(withoutReasons.NewEvalEvent(flag1v1, user, ldreason.NewEvaluationDetail(ldvalue.String("b"), 2, noReason),
			flag1Default, ""))
		es.summarizeEvent(withoutReasons.NewEvalEvent(flag1v1, user, ldreason.NewEvaluationDetail(ldvalue.String("a"), 1, noReason),
			flag1Default, ""))
		es.summarizeEvent(withoutReasons.NewEvalEvent(flag1v2, user, ldreason.NewEvaluationDetail(ldvalue.String("a"), 1, noReason),
			flag1Default, ""))
		es.summarizeEvent(withoutReasons.NewEvalEvent(flag2, user, ldreason.NewEvaluationDetail(ldvalue.String("c"), 3, noReason),
			flag2Default, ""))

		bytes, count := defaultFormatter.makeOutputEvents(nil, es.snapshot())
		require.Equal(t, 1, count)
		outValue := ldvalue.Parse(bytes).GetByIndex(0)
		require.NotEqual(t, ldvalue.Null(), outValue)
		featuresObj := outValue.GetByKey("features")
		assert.Equal(t, 2, featuresObj.Count())

		flag1Obj := featuresObj.GetByKey("flag1")
		assert.Equal(t, flag1Default, flag1Obj.GetByKey("default"))
		flag1Counters := getArrayValues(flag1Obj.GetByKey("counters"))
		assert.Len(t, flag1Counters, 3)
		sort.Slice(flag1Counters, func(i, j int) bool {
			iVers := flag1Counters[i].GetByKey("version").IntValue()
			jVers := flag1Counters[j].GetByKey("version").IntValue()
			if iVers != jVers {
				return iVers < jVers
			}
			iVar := flag1Counters[i].GetByKey("variation").IntValue()
			jVar := flag1Counters[j].GetByKey("variation").IntValue()
			return iVar < jVar
		})
		assert.Equal(t, ldvalue.Parse([]byte(`{"version":100,"variation":1,"value":"a","count":2}`)),
			flag1Counters[0])
		assert.Equal(t, ldvalue.Parse([]byte(`{"version":100,"variation":2,"value":"b","count":1}`)),
			flag1Counters[1])
		assert.Equal(t, ldvalue.Parse([]byte(`{"version":200,"variation":1,"value":"a","count":1}`)),
			flag1Counters[2])

		flag2Obj := featuresObj.GetByKey("flag2")
		assert.Equal(t, flag2Default, flag2Obj.GetByKey("default"))
		flag2Counters := getArrayValues(flag2Obj.GetByKey("counters"))
		assert.Len(t, flag2Counters, 1)
		assert.Equal(t, ldvalue.Parse([]byte(`{"version":1,"variation":3,"value":"c","count":1}`)),
			flag2Counters[0])
	})

	t.Run("empty payload", func(t *testing.T) {
		bytes, count := defaultFormatter.makeOutputEvents([]Event{}, eventSummary{})
		assert.Nil(t, bytes)
		assert.Equal(t, 0, count)
	})
}

func verifyEventOutput(t *testing.T, formatter eventOutputFormatter, event Event, expectedJSON string) {
	expectedValue := ldvalue.Parse([]byte(expectedJSON))
	bytes, count := formatter.makeOutputEvents([]Event{event}, eventSummary{})
	require.Equal(t, 1, count)
	outValue := ldvalue.Parse(bytes)
	require.Equal(t, outValue.Count(), 1)
	outEvent := outValue.GetByIndex(0)
	assert.Equal(t, expectedValue, outEvent, "expected JSON: %s, actual JSON: %s", expectedJSON, outEvent)
}

func verifySummaryEventOutput(t *testing.T, formatter eventOutputFormatter, summary eventSummary, expectedJSON string) {
	expectedValue := ldvalue.Parse([]byte(expectedJSON))
	bytes, count := formatter.makeOutputEvents(nil, summary)
	require.Equal(t, 1, count)
	outValue := ldvalue.Parse(bytes)
	require.Equal(t, outValue.Count(), 1)
	outEvent := outValue.GetByIndex(0)
	assert.Equal(t, expectedValue, outEvent, "expected JSON: %s, actual JSON: %s", expectedJSON, outEvent)
}

func getArrayValues(v ldvalue.Value) []ldvalue.Value {
	s := make([]ldvalue.Value, 0, v.Count())
	v.Enumerate(func(i int, k string, vv ldvalue.Value) bool {
		s = append(s, vv)
		return true
	})
	return s
}
