package ldmodel

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

func makeClause(attr lduser.UserAttribute, op Operator, values ...ldvalue.Value) *Clause {
	return &Clause{Attribute: attr, Op: op, Values: values}
}

func TestClauseCanMatchBuiltInAttribute(t *testing.T) {
	clause := makeClause(lduser.NameAttribute, OperatorIn, ldvalue.String("Bob"))
	user := lduser.NewUserBuilder("key").Name("Bob").Build()
	assert.True(t, ClauseMatchesUser(clause, &user))
}

func TestClauseCanMatchCustomAttribute(t *testing.T) {
	clause := makeClause(lduser.UserAttribute("legs"), OperatorIn, ldvalue.Int(4))
	user := lduser.NewUserBuilder("key").Custom("legs", ldvalue.Int(4)).Build()
	assert.True(t, ClauseMatchesUser(clause, &user))
}

func TestClauseReturnsFalseForMissingAttribute(t *testing.T) {
	clause := makeClause(lduser.UserAttribute("legs"), OperatorIn, ldvalue.Int(4))
	user := lduser.NewUserBuilder("key").Name("Bob").Build()
	assert.False(t, ClauseMatchesUser(clause, &user))
}

func TestClauseMatchesIfAnyClauseValueMatches(t *testing.T) {
	clause := makeClause(lduser.KeyAttribute, OperatorIn, ldvalue.String("key1"), ldvalue.String("key2"))
	user := lduser.NewUser("key2")
	assert.True(t, ClauseMatchesUser(clause, &user))
}

func TestClauseMatchesIfAnyElementInUserArrayValueMatchesAnyClauseValue(t *testing.T) {
	clause := makeClause(lduser.UserAttribute("pets"), OperatorIn, ldvalue.String("cat"), ldvalue.String("dog"))
	user := lduser.NewUserBuilder("key").Custom("pets", ldvalue.ArrayOf(ldvalue.String("fish"), ldvalue.String("dog"))).Build()
	assert.True(t, ClauseMatchesUser(clause, &user))
}

func TestClauseDoesNotMatchIfNoElementInUserArrayValueMatchesAnyClauseValue(t *testing.T) {
	clause := makeClause(lduser.UserAttribute("pets"), OperatorIn, ldvalue.String("cat"), ldvalue.String("dog"))
	user := lduser.NewUserBuilder("key").Custom("pets", ldvalue.ArrayOf(ldvalue.String("fish"), ldvalue.String("bird"))).Build()
	assert.False(t, ClauseMatchesUser(clause, &user))
}

func TestClauseCanBeNegated(t *testing.T) {
	clause := makeClause(lduser.NameAttribute, OperatorIn, ldvalue.String("Bob"))
	clause.Negate = true
	user := lduser.NewUserBuilder("key").Name("Bob").Build()
	assert.False(t, ClauseMatchesUser(clause, &user))
}

func TestClauseForMissingAttributeIsFalseEvenIfNegated(t *testing.T) {
	clause := makeClause(lduser.UserAttribute("legs"), OperatorIn, ldvalue.Int(4))
	clause.Negate = true
	user := lduser.NewUserBuilder("key").Name("Bob").Build()
	assert.False(t, ClauseMatchesUser(clause, &user))
}

func TestClauseWithUnknownOperatorDoesNotMatch(t *testing.T) {
	clause := makeClause(lduser.NameAttribute, "doesSomethingUnsupported", ldvalue.String("Bob"))
	user := lduser.NewUserBuilder("key").Name("Bob").Build()
	assert.False(t, ClauseMatchesUser(clause, &user))
}
