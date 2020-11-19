package evaluation

import (
	"fmt"

	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
	"gopkg.in/launchdarkly/go-server-sdk-evaluation.v1/ldbuilders"
	"gopkg.in/launchdarkly/go-server-sdk-evaluation.v1/ldmodel"
)

type simpleDataProvider struct {
	getFlag    func(string) *ldmodel.FeatureFlag
	getSegment func(string) *ldmodel.Segment
}

func (s *simpleDataProvider) GetFeatureFlag(key string) *ldmodel.FeatureFlag {
	return s.getFlag(key)
}

func (s *simpleDataProvider) GetSegment(key string) *ldmodel.Segment {
	return s.getSegment(key)
}

func (s *simpleDataProvider) withStoredFlags(flags ...ldmodel.FeatureFlag) *simpleDataProvider {
	return &simpleDataProvider{
		getFlag: func(key string) *ldmodel.FeatureFlag {
			for _, f := range flags {
				if f.Key == key {
					ff := f
					return &ff
				}
			}
			return s.getFlag(key)
		},
		getSegment: s.getSegment,
	}
}

func (s *simpleDataProvider) withNonexistentFlag(flagKey string) *simpleDataProvider {
	return &simpleDataProvider{
		getFlag: func(key string) *ldmodel.FeatureFlag {
			if key == flagKey {
				return nil
			}
			return s.getFlag(key)
		},
		getSegment: s.getSegment,
	}
}

func (s *simpleDataProvider) withStoredSegments(segments ...ldmodel.Segment) *simpleDataProvider {
	return &simpleDataProvider{
		getFlag: s.getFlag,
		getSegment: func(key string) *ldmodel.Segment {
			for _, seg := range segments {
				if seg.Key == key {
					ss := seg
					return &ss
				}
			}
			return s.getSegment(key)
		},
	}
}

func (s *simpleDataProvider) withNonexistentSegment(segmentKey string) *simpleDataProvider {
	return &simpleDataProvider{
		getFlag: s.getFlag,
		getSegment: func(key string) *ldmodel.Segment {
			if key == segmentKey {
				return nil
			}
			return s.getSegment(key)
		},
	}
}

func basicDataProvider() *simpleDataProvider {
	return &simpleDataProvider{
		getFlag: func(key string) *ldmodel.FeatureFlag {
			panic(fmt.Errorf("unexpectedly queried feature flag: %s", key))
		},
		getSegment: func(key string) *ldmodel.Segment {
			panic(fmt.Errorf("unexpectedly queried segment: %s", key))
		},
	}
}

func basicEvaluator() Evaluator {
	return NewEvaluator(basicDataProvider())
}

type prereqEventSink struct {
	events []PrerequisiteFlagEvent
}

func (p *prereqEventSink) record(event PrerequisiteFlagEvent) {
	p.events = append(p.events, event)
}

func makeClauseToMatchUser(user lduser.User) ldmodel.Clause {
	return ldbuilders.Clause(lduser.KeyAttribute, ldmodel.OperatorIn, ldvalue.String(user.GetKey()))
}

func makeClauseToNotMatchUser(user lduser.User) ldmodel.Clause {
	return ldbuilders.Clause(lduser.KeyAttribute, ldmodel.OperatorIn, ldvalue.String("not-"+user.GetKey()))
}

func makeFlagToMatchUser(user lduser.User, variationOrRollout ldmodel.VariationOrRollout) ldmodel.FeatureFlag {
	return ldbuilders.NewFlagBuilder("feature").
		On(true).
		OffVariation(1).
		AddRule(ldbuilders.NewRuleBuilder().ID("rule-id").VariationOrRollout(variationOrRollout).Clauses(makeClauseToMatchUser(user))).
		FallthroughVariation(0).
		Variations(fallthroughValue, offValue, onValue).
		Build()
}

func booleanFlagWithClause(clause ldmodel.Clause) ldmodel.FeatureFlag {
	return ldbuilders.NewFlagBuilder("feature").
		On(true).
		AddRule(ldbuilders.NewRuleBuilder().Variation(1).Clauses(clause)).
		Variations(ldvalue.Bool(false), ldvalue.Bool(true)).
		FallthroughVariation(0).
		Build()
}

func booleanFlagWithSegmentMatch(segmentKeys ...string) ldmodel.FeatureFlag {
	return booleanFlagWithClause(ldbuilders.SegmentMatchClause(segmentKeys...))
}
