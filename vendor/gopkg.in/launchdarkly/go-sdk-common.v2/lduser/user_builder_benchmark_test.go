package lduser

import (
	"testing"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

var (
	benchmarkUserBuilder                        = NewUserBuilder("key")
	benchmarkUserBuilderWithAllScalarAttributes = NewUserBuilder("key").
							Secondary("s").
							IP("i").
							Country("c").
							Email("e").
							FirstName("f").
							LastName("l").
							Avatar("a").
							Name("n").
							Anonymous(true)

	benchmarkUserResult User
)

func BenchmarkNewUserNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkUserResult = NewUser("key")
	}
}

func BenchmarkNewAnonymousUserNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkUserResult = NewAnonymousUser("key")
	}
}

func BenchmarkNewUserFromBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkUserResult = NewUserBuilder("key").Name("name").Email("email").Build()
	}
}

func BenchmarkNewUserFromBuilderWithCustom(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkUserResult = NewUserBuilder("key").
			Custom("attr1", ldvalue.String("value1")).
			Custom("attr2", ldvalue.String("value2")).
			Build()
	}
}

func BenchmarkNewUserFromBuilderWithPrivate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkUserResult = NewUserBuilder("key").Name("name").AsPrivateAttribute().Build()
	}
}

func BenchmarkBuildUserWithAllScalarAttributesNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkUserResult = benchmarkUserBuilderWithAllScalarAttributes.Build()
	}
}

func BenchmarkUserBuilderSetScalarAttributesNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkUserBuilder.
			Secondary("s").
			IP("i").
			Country("c").
			Email("e").
			FirstName("f").
			LastName("l").
			Avatar("a").
			Name("n").
			Anonymous(true)
	}
}
