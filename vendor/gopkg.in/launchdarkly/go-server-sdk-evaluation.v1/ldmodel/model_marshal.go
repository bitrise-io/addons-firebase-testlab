package ldmodel

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream"
)

// For backward compatibility, we are only allowed to drop out properties that have default values if
// Go SDK 4.x would also have done so (since some SDKs are not tolerant of missing properties in
// general). This is true of all properties that have OptionalInt-like semantics (having either a
// numeric value or "undefined"), and properties that could be either a JSON object or null (like
// VariationOrRollout.Rollout), and the BucketBy property which has optional-string-like behavior.
// Array properties should not be dropped even if nil.
//
// Properties that did not exist prior to Go SDK v5 are always safe to drop if they have default
// values, since older SDKs will never look for them. These are:
// - FeatureFlag.ClientSideAvailability
// - Segment.Unbounded

func marshalFeatureFlag(flag FeatureFlag) ([]byte, error) {
	var b jsonstream.JSONBuffer
	b.Grow(200)

	b.BeginObject()

	writeString(&b, "key", flag.Key)

	writeBool(&b, "on", flag.On)

	b.WriteName("prerequisites")
	b.BeginArray()
	for _, p := range flag.Prerequisites {
		b.BeginObject()
		writeString(&b, "key", p.Key)
		writeInt(&b, "variation", p.Variation)
		b.EndObject()
	}
	b.EndArray()

	b.WriteName("targets")
	b.BeginArray()
	for _, t := range flag.Targets {
		b.BeginObject()
		writeInt(&b, "variation", t.Variation)
		writeStringArray(&b, "values", t.Values)
		b.EndObject()
	}
	b.EndArray()

	b.WriteName("rules")
	b.BeginArray()
	for _, r := range flag.Rules {
		b.BeginObject()
		writeVariationOrRolloutProperties(&b, r.VariationOrRollout)
		if r.ID != "" {
			writeString(&b, "id", r.ID)
		}
		writeClauses(&b, r.Clauses)
		writeBool(&b, "trackEvents", r.TrackEvents)
		b.EndObject()
	}
	b.EndArray()

	b.WriteName("fallthrough")
	b.BeginObject()
	writeVariationOrRolloutProperties(&b, flag.Fallthrough)
	b.EndObject()

	b.WriteName("offVariation")
	flag.OffVariation.WriteToJSONBuffer(&b)

	b.WriteName("variations")
	b.BeginArray()
	for _, v := range flag.Variations {
		v.WriteToJSONBuffer(&b)
	}
	b.EndArray()

	// In the older JSON schema, ClientSideAvailability.UsingEnvironmentID was in "clientSide", and
	// ClientSideAvailability.UsingMobileKey was assumed to be true. In the newer schema, those are
	// both in a "clientSideAvailability" object.
	//
	// If ClientSideAvailability.Explicit is true, then this flag used the newer schema and should be
	// reserialized the same way. If it is false, we will reserialize with the old schema, which
	// does not include UsingMobileKey; note that in that case UsingMobileKey is assumed to be true.
	//
	// For backward compatibility with older SDKs that might be reading a flag that was serialized by
	// this SDK, we always include the older "clientSide" property if it would be true.
	if flag.ClientSideAvailability.Explicit {
		b.WriteName("clientSideAvailability")
		b.BeginObject()
		b.WriteName("usingMobileKey")
		b.WriteBool(flag.ClientSideAvailability.UsingMobileKey)
		b.WriteName("usingEnvironmentId")
		b.WriteBool(flag.ClientSideAvailability.UsingEnvironmentID)
		b.EndObject()
	}
	writeBool(&b, "clientSide", flag.ClientSideAvailability.UsingEnvironmentID)

	writeString(&b, "salt", flag.Salt)

	writeBool(&b, "trackEvents", flag.TrackEvents)
	writeBool(&b, "trackEventsFallthrough", flag.TrackEventsFallthrough)

	b.WriteName("debugEventsUntilDate")
	if flag.DebugEventsUntilDate == 0 {
		b.WriteNull()
	} else {
		b.WriteUint64(uint64(flag.DebugEventsUntilDate))
	}

	writeInt(&b, "version", flag.Version)

	writeBool(&b, "deleted", flag.Deleted)

	b.EndObject()

	return b.Get()
}

func marshalSegment(segment Segment) ([]byte, error) {
	var b jsonstream.JSONBuffer
	b.Grow(200)

	b.BeginObject()

	writeString(&b, "key", segment.Key)
	writeStringArray(&b, "included", segment.Included)
	writeStringArray(&b, "excluded", segment.Excluded)
	writeString(&b, "salt", segment.Salt)

	b.WriteName("rules")
	b.BeginArray()
	for _, r := range segment.Rules {
		b.BeginObject()
		writeString(&b, "id", r.ID)
		writeClauses(&b, r.Clauses)
		if r.Weight >= 0 {
			writeInt(&b, "weight", r.Weight)
		}
		if r.BucketBy != "" {
			writeString(&b, "bucketBy", string(r.BucketBy))
		}
		b.EndObject()
	}
	b.EndArray()

	if segment.Unbounded {
		writeBool(&b, "unbounded", segment.Unbounded)
	}

	writeInt(&b, "version", segment.Version)
	writeBool(&b, "deleted", segment.Deleted)

	b.EndObject()
	return b.Get()
}

func writeBool(b *jsonstream.JSONBuffer, name string, value bool) {
	b.WriteName(name)
	b.WriteBool(value)
}

func writeInt(b *jsonstream.JSONBuffer, name string, value int) {
	b.WriteName(name)
	b.WriteInt(value)
}

func writeString(b *jsonstream.JSONBuffer, name string, value string) {
	b.WriteName(name)
	b.WriteString(value)
}

func writeStringArray(b *jsonstream.JSONBuffer, name string, values []string) {
	b.WriteName(name)
	b.BeginArray()
	for _, v := range values {
		b.WriteString(v)
	}
	b.EndArray()
}

func writeVariationOrRolloutProperties(b *jsonstream.JSONBuffer, vr VariationOrRollout) {
	if vr.Variation.IsDefined() {
		writeInt(b, "variation", vr.Variation.IntValue())
	}
	if len(vr.Rollout.Variations) > 0 {
		b.WriteName("rollout")
		b.BeginObject()
		b.WriteName("variations")
		b.BeginArray()
		for _, wv := range vr.Rollout.Variations {
			b.BeginObject()
			writeInt(b, "variation", wv.Variation)
			writeInt(b, "weight", wv.Weight)
			b.EndObject()
		}
		b.EndArray()
		if vr.Rollout.BucketBy != "" {
			writeString(b, "bucketBy", string(vr.Rollout.BucketBy))
		}
		b.EndObject()
	}
}

func writeClauses(b *jsonstream.JSONBuffer, clauses []Clause) {
	b.WriteName("clauses")
	b.BeginArray()
	for _, c := range clauses {
		b.BeginObject()
		writeString(b, "attribute", string(c.Attribute))
		writeString(b, "op", string(c.Op))
		b.WriteName("values")
		b.BeginArray()
		for _, v := range c.Values {
			v.WriteToJSONBuffer(b)
		}
		b.EndArray()
		writeBool(b, "negate", c.Negate)
		b.EndObject()
	}
	b.EndArray()
}
