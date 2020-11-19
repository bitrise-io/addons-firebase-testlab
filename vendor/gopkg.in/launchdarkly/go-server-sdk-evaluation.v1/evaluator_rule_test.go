package evaluation

import (
	"testing"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldreason"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
	"gopkg.in/launchdarkly/go-server-sdk-evaluation.v1/ldbuilders"
	"gopkg.in/launchdarkly/go-server-sdk-evaluation.v1/ldmodel"

	"github.com/stretchr/testify/assert"
)

func TestRuleWithTooHighVariationIndexReturnsMalformedFlagError(t *testing.T) {
	user := lduser.NewUser("userkey")
	f := makeFlagToMatchUser(user, ldbuilders.Variation(999))

	eventSink := prereqEventSink{}
	result := basicEvaluator().Evaluate(&f, user, eventSink.record)
	assert.Equal(t, ldreason.NewEvaluationDetailForError(ldreason.EvalErrorMalformedFlag, ldvalue.Null()), result)
	assert.Equal(t, 0, len(eventSink.events))
}

func TestRuleWithNegativeVariationIndexReturnsMalformedFlagError(t *testing.T) {
	user := lduser.NewUser("userkey")
	f := makeFlagToMatchUser(user, ldbuilders.Variation(-1))

	eventSink := prereqEventSink{}
	result := basicEvaluator().Evaluate(&f, user, eventSink.record)
	assert.Equal(t, ldreason.NewEvaluationDetailForError(ldreason.EvalErrorMalformedFlag, ldvalue.Null()), result)
	assert.Equal(t, 0, len(eventSink.events))
}

func TestRuleWithNoVariationOrRolloutReturnsMalformedFlagError(t *testing.T) {
	user := lduser.NewUser("userkey")
	f := makeFlagToMatchUser(user, ldbuilders.Rollout())

	eventSink := prereqEventSink{}
	result := basicEvaluator().Evaluate(&f, user, eventSink.record)
	assert.Equal(t, ldreason.NewEvaluationDetailForError(ldreason.EvalErrorMalformedFlag, ldvalue.Null()), result)
	assert.Equal(t, 0, len(eventSink.events))
}

func TestClauseWithUnknownOperatorDoesNotStopSubsequentRuleFromMatching(t *testing.T) {
	badClause := ldbuilders.Clause(lduser.NameAttribute, "doesSomethingUnsupported", ldvalue.String("Bob"))
	goodClause := ldbuilders.Clause(lduser.NameAttribute, ldmodel.OperatorIn, ldvalue.String("Bob"))
	f := ldbuilders.NewFlagBuilder("feature").
		On(true).
		AddRule(ldbuilders.NewRuleBuilder().ID("bad").Variation(1).Clauses(badClause)).
		AddRule(ldbuilders.NewRuleBuilder().ID("good").Variation(1).Clauses(goodClause)).
		Variations(ldvalue.Bool(false), ldvalue.Bool(true)).
		Build()
	user := lduser.NewUserBuilder("key").Name("Bob").Build()

	result := basicEvaluator().Evaluate(&f, user, nil)
	assert.True(t, result.Value.BoolValue())
	assert.Equal(t, ldreason.NewEvalReasonRuleMatch(1, "good"), result.Reason)
}
