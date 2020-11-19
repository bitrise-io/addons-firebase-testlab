package ldevents

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldtime"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

// Manages the state of summarizable information for the EventProcessor, including the
// event counters and user deduplication. Note that the methods for this type are
// deliberately not thread-safe, because they should always be called from EventProcessor's
// single event-processing goroutine.
type eventSummarizer struct {
	eventsState eventSummary
}

type eventSummary struct {
	counters  map[counterKey]*counterValue
	startDate ldtime.UnixMillisecondTime
	endDate   ldtime.UnixMillisecondTime
}

type counterKey struct {
	key       string
	variation ldvalue.OptionalInt
	version   ldvalue.OptionalInt
}

type counterValue struct {
	count       int
	flagValue   ldvalue.Value
	flagDefault ldvalue.Value
}

func newEventSummarizer() eventSummarizer {
	return eventSummarizer{eventsState: newEventSummary()}
}

func newEventSummary() eventSummary {
	return eventSummary{
		counters: make(map[counterKey]*counterValue),
	}
}

// Adds this event to our counters.
func (s *eventSummarizer) summarizeEvent(fe FeatureRequestEvent) {
	key := counterKey{key: fe.Key, variation: fe.Variation, version: fe.Version}

	if value, ok := s.eventsState.counters[key]; ok {
		value.count++
	} else {
		s.eventsState.counters[key] = &counterValue{
			count:       1,
			flagValue:   fe.Value,
			flagDefault: fe.Default,
		}
	}

	creationDate := fe.CreationDate
	if s.eventsState.startDate == 0 || creationDate < s.eventsState.startDate {
		s.eventsState.startDate = creationDate
	}
	if creationDate > s.eventsState.endDate {
		s.eventsState.endDate = creationDate
	}
}

// Returns a snapshot of the current summarized event data.
func (s *eventSummarizer) snapshot() eventSummary {
	return s.eventsState
}

func (s *eventSummarizer) reset() {
	s.eventsState = newEventSummary()
}
