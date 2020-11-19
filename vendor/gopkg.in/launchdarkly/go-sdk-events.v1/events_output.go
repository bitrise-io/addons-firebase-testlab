package ldevents

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldtime"
)

// In this file we create the analytics event JSON data that we send to LaunchDarkly. This does not
// always correspond to the shape of the event objects that are fed into EventProcessor.

// Event types
const (
	FeatureRequestEventKind = "feature"
	FeatureDebugEventKind   = "debug"
	CustomEventKind         = "custom"
	IdentifyEventKind       = "identify"
	IndexEventKind          = "index"
	SummaryEventKind        = "summary"
)

type eventOutputFormatter struct {
	userFilter userFilter
	config     EventsConfiguration
}

func (ef eventOutputFormatter) makeOutputEvents(events []Event, summary eventSummary) ([]byte, int) {
	n := len(events)
	var b jsonstream.JSONBuffer
	b.BeginArray()

	for _, e := range events {
		ef.writeOutputEvent(&b, e)
	}
	if len(summary.counters) > 0 {
		ef.writeSummaryEvent(&b, summary)
		n++
	}

	if n > 0 {
		b.EndArray()
		bytes, _ := b.Get()
		return bytes, n
	}
	return nil, 0
}

func (ef eventOutputFormatter) writeOutputEvent(b *jsonstream.JSONBuffer, evt Event) {
	b.BeginObject()

	switch evt := evt.(type) {
	case FeatureRequestEvent:
		kind := FeatureRequestEventKind
		if evt.Debug {
			kind = FeatureDebugEventKind
		}
		beginEventFields(b, kind, evt.BaseEvent.CreationDate)
		b.WriteName("key")
		b.WriteString(evt.Key)
		if evt.Version.IsDefined() {
			b.WriteName("version")
			b.WriteInt(evt.Version.IntValue())
		}
		if ef.config.InlineUsersInEvents || evt.Debug {
			b.WriteName("user")
			ef.userFilter.writeUser(b, evt.User)
		} else {
			b.WriteName("userKey")
			b.WriteString(evt.User.GetKey())
		}
		if evt.Variation.IsDefined() {
			b.WriteName("variation")
			b.WriteInt(evt.Variation.IntValue())
		}
		b.WriteName("value")
		evt.Value.WriteToJSONBuffer(b)
		b.WriteName("default")
		evt.Default.WriteToJSONBuffer(b)
		if pre, ok := evt.PrereqOf.Get(); ok {
			b.WriteName("prereqOf")
			b.WriteString(pre)
		}
		if evt.Reason.GetKind() != "" {
			b.WriteName("reason")
			evt.Reason.WriteToJSONBuffer(b)
		}

	case CustomEvent:
		beginEventFields(b, CustomEventKind, evt.BaseEvent.CreationDate)
		b.WriteName("key")
		b.WriteString(evt.Key)
		if !evt.Data.IsNull() {
			b.WriteName("data")
			evt.Data.WriteToJSONBuffer(b)
		}
		if ef.config.InlineUsersInEvents {
			b.WriteName("user")
			ef.userFilter.writeUser(b, evt.User)
		} else {
			b.WriteName("userKey")
			b.WriteString(evt.User.GetKey())
		}
		if evt.HasMetric {
			b.WriteName("metricValue")
			b.WriteFloat64(evt.MetricValue)
		}

	case IdentifyEvent:
		beginEventFields(b, IdentifyEventKind, evt.BaseEvent.CreationDate)
		b.WriteName("key")
		b.WriteString(evt.User.GetKey())
		b.WriteName("user")
		ef.userFilter.writeUser(b, evt.User)

	case indexEvent:
		beginEventFields(b, IndexEventKind, evt.BaseEvent.CreationDate)
		b.WriteName("user")
		ef.userFilter.writeUser(b, evt.User)
	}

	b.EndObject()
}

func beginEventFields(b *jsonstream.JSONBuffer, kind string, creationDate ldtime.UnixMillisecondTime) {
	b.WriteName("kind")
	b.WriteString(kind)
	b.WriteName("creationDate")
	b.WriteUint64(uint64(creationDate))
}

// Transforms the summary data into the format used for event sending.
func (ef eventOutputFormatter) writeSummaryEvent(b *jsonstream.JSONBuffer, snapshot eventSummary) {
	b.BeginObject()

	b.WriteName("kind")
	b.WriteString(SummaryEventKind)
	b.WriteName("startDate")
	b.WriteUint64(uint64(snapshot.startDate))
	b.WriteName("endDate")
	b.WriteUint64(uint64(snapshot.endDate))

	b.WriteName("features")
	b.BeginObject()

	// snapshot.counters contains composite keys in any order, part of which is the flag key; we want to group
	// them together by flag key. The following multi-pass logic allows us to do this without using maps. It is
	// based on the corresponding Java SDK logic in EventOutputFormatter.java.

	unprocessedKeys := make([]counterKey, 0, 100)
	// starting with a fixed capacity allows this slice to live on the stack unless it grows past that limit
	for key := range snapshot.counters {
		unprocessedKeys = append(unprocessedKeys, key)
	}

	for i, key := range unprocessedKeys {
		if key.key == "" {
			continue // we've already consumed this one when we saw the same flag key further up
		}
		// Now we've got a flag key that we haven't seen before
		flagKey := key.key
		firstValue := snapshot.counters[key]

		b.WriteName(flagKey)
		b.BeginObject()

		b.WriteName("default")
		firstValue.flagDefault.WriteToJSONBuffer(b)

		b.WriteName("counters")
		b.BeginArray()

		// Iterate over the remainder of the list to find any more entries for the same flag key.
		for j := i; j < len(unprocessedKeys); j++ {
			var anotherKey counterKey
			var anotherValue *counterValue
			if j == i {
				anotherKey = key
				anotherValue = firstValue
			} else {
				anotherKey = unprocessedKeys[j]
				if anotherKey.key != flagKey {
					continue
				}
				anotherValue = snapshot.counters[anotherKey]
				unprocessedKeys[j].key = "" // clear this one so we'll skip it in the next pass
			}

			b.BeginObject()
			if anotherKey.variation.IsDefined() {
				b.WriteName("variation")
				b.WriteInt(anotherKey.variation.IntValue())
			}
			if anotherKey.version.IsDefined() {
				b.WriteName("version")
				b.WriteInt(anotherKey.version.IntValue())
			} else {
				b.WriteName("unknown")
				b.WriteBool(true)
			}
			b.WriteName("value")
			anotherValue.flagValue.WriteToJSONBuffer(b)
			b.WriteName("count")
			b.WriteInt(anotherValue.count)
			b.EndObject() // end of this counter
		}

		b.EndArray()  // end of "counters" array
		b.EndObject() // end of this flag
	}
	b.EndObject() // end of "features"
	b.EndObject() // end of summary event
}
