package ldmodel

import (
	"encoding/json"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldtime"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

// These "JSONRep" types are for use with json.Unmarshal. We will translate them to our real data model types.
// The differences are due to having to use pointers to represent optional values with json.Unmarshal; we do
// not want to use pointers in the internal model due to their safety issues.

type featureFlagJSONRep struct {
	Key                    string                         `json:"key"`
	On                     bool                           `json:"on"`
	Prerequisites          []prerequisiteJSONRep          `json:"prerequisites"`
	Targets                []targetJSONRep                `json:"targets"`
	Rules                  []flagRuleJSONRep              `json:"rules"`
	Fallthrough            variationOrRolloutJSONRep      `json:"fallthrough"`
	OffVariation           ldvalue.OptionalInt            `json:"offVariation"`
	Variations             []ldvalue.Value                `json:"variations"`
	ClientSide             bool                           `json:"clientSide"`
	ClientSideAvailability *clientSideAvailabilityJSONRep `json:"clientSideAvailability"`
	Salt                   string                         `json:"salt"`
	TrackEvents            bool                           `json:"trackEvents"`
	TrackEventsFallthrough bool                           `json:"trackEventsFallthrough"`
	DebugEventsUntilDate   ldtime.UnixMillisecondTime     `json:"debugEventsUntilDate"`
	Version                int                            `json:"version"`
	Deleted                bool                           `json:"deleted"`
}

type prerequisiteJSONRep struct {
	Key       string `json:"key"`
	Variation int    `json:"variation"`
}

type targetJSONRep struct {
	Values    []string `json:"values"`
	Variation int      `json:"variation"`
}

type flagRuleJSONRep struct {
	variationOrRolloutJSONRep
	ID          string          `json:"id"`
	Clauses     []clauseJSONRep `json:"clauses"`
	TrackEvents bool            `json:"trackEvents"`
}

type clauseJSONRep struct {
	Attribute lduser.UserAttribute `json:"attribute"`
	Op        Operator             `json:"op"`
	Values    []ldvalue.Value      `json:"values" bson:"values"` // An array, interpreted as an OR of values
	Negate    bool                 `json:"negate"`
}

type variationOrRolloutJSONRep struct {
	Variation ldvalue.OptionalInt `json:"variation"`
	Rollout   *rolloutJSONRep     `json:"rollout"`
}

type rolloutJSONRep struct {
	Variations []weightedVariationJSONRep `json:"variations"`
	BucketBy   lduser.UserAttribute       `json:"bucketBy"`
}

type weightedVariationJSONRep struct {
	Variation int `json:"variation"`
	Weight    int `json:"weight"`
}

type segmentJSONRep struct {
	Key       string               `json:"key"`
	Included  []string             `json:"included"`
	Excluded  []string             `json:"excluded"`
	Salt      string               `json:"salt"`
	Rules     []segmentRuleJSONRep `json:"rules"`
	Unbounded bool                 `json:"unbounded"`
	Version   int                  `json:"version"`
	Deleted   bool                 `json:"deleted"`
}

type segmentRuleJSONRep struct {
	ID       string                `json:"id"`
	Clauses  []clauseJSONRep       `json:"clauses"`
	Weight   *int                  `json:"weight"`
	BucketBy *lduser.UserAttribute `json:"bucketBy"`
}

type clientSideAvailabilityJSONRep struct {
	UsingMobileKey     bool `json:"usingMobileKey"`
	UsingEnvironmentID bool `json:"usingEnvironmentId"`
}

func unmarshalFeatureFlag(data []byte) (FeatureFlag, error) {
	var fields featureFlagJSONRep
	if err := json.Unmarshal(data, &fields); err != nil {
		return FeatureFlag{}, err
	}

	ret := FeatureFlag{
		Key:          fields.Key,
		Version:      fields.Version,
		Deleted:      fields.Deleted,
		On:           fields.On,
		OffVariation: fields.OffVariation,
		Variations:   fields.Variations,
	}
	if len(fields.Prerequisites) > 0 {
		ret.Prerequisites = make([]Prerequisite, len(fields.Prerequisites))
		for i, p := range fields.Prerequisites {
			ret.Prerequisites[i] = Prerequisite(p) // fields are the same
		}
	}
	if len(fields.Targets) > 0 {
		ret.Targets = make([]Target, len(fields.Targets))
		for i, t := range fields.Targets {
			ret.Targets[i] = Target{
				Values:    t.Values,
				Variation: t.Variation,
			}
		}
	}
	if len(fields.Rules) > 0 {
		ret.Rules = make([]FlagRule, len(fields.Rules))
		for i, r := range fields.Rules {
			fr := FlagRule{
				VariationOrRollout: decodeVariationOrRollout(r.variationOrRolloutJSONRep),
				ID:                 r.ID,
				Clauses:            decodeClauses(r.Clauses),
				TrackEvents:        r.TrackEvents,
			}
			ret.Rules[i] = fr
		}
	}
	ret.Fallthrough = decodeVariationOrRollout(fields.Fallthrough)
	if fields.ClientSideAvailability == nil {
		ret.ClientSideAvailability = ClientSideAvailability{
			UsingMobileKey:     true, // always assumed to be true in the old schema
			UsingEnvironmentID: fields.ClientSide,
			Explicit:           false,
		}
	} else {
		ret.ClientSideAvailability = ClientSideAvailability{
			UsingMobileKey:     fields.ClientSideAvailability.UsingMobileKey,
			UsingEnvironmentID: fields.ClientSideAvailability.UsingEnvironmentID,
			Explicit:           true,
		}
	}
	ret.Salt = fields.Salt
	ret.TrackEvents = fields.TrackEvents
	ret.TrackEventsFallthrough = fields.TrackEventsFallthrough
	ret.DebugEventsUntilDate = fields.DebugEventsUntilDate

	PreprocessFlag(&ret)
	return ret, nil
}

func unmarshalSegment(data []byte) (Segment, error) {
	var fields segmentJSONRep
	if err := json.Unmarshal(data, &fields); err != nil {
		return Segment{}, err
	}

	ret := Segment{
		Key:       fields.Key,
		Version:   fields.Version,
		Deleted:   fields.Deleted,
		Included:  fields.Included,
		Excluded:  fields.Excluded,
		Salt:      fields.Salt,
		Unbounded: fields.Unbounded,
	}
	ret.Rules = make([]SegmentRule, len(fields.Rules))
	for i, r := range fields.Rules {
		sr := SegmentRule{
			ID:      r.ID,
			Clauses: decodeClauses(r.Clauses),
		}
		if r.Weight == nil {
			sr.Weight = -1
		} else {
			sr.Weight = *r.Weight
		}
		if r.BucketBy != nil {
			sr.BucketBy = *r.BucketBy
		}
		ret.Rules[i] = sr
	}

	PreprocessSegment(&ret)
	return ret, nil
}

func decodeVariationOrRollout(fields variationOrRolloutJSONRep) VariationOrRollout {
	ret := VariationOrRollout{Variation: fields.Variation}
	if fields.Rollout != nil {
		ret.Rollout.Variations = make([]WeightedVariation, len(fields.Rollout.Variations))
		for i, wv := range fields.Rollout.Variations {
			ret.Rollout.Variations[i] = WeightedVariation(wv) // fields are the same
		}
		ret.Rollout.BucketBy = fields.Rollout.BucketBy
	}
	return ret
}

func decodeClauses(clauses []clauseJSONRep) []Clause {
	ret := make([]Clause, len(clauses))
	for i, c := range clauses {
		ret[i] = Clause{
			Attribute: c.Attribute,
			Op:        c.Op,
			Values:    c.Values,
			Negate:    c.Negate,
		}
	}
	return ret
}
