package lduser

import (
	"encoding/json"
	"testing"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

var (
	benchmarkSimpleUser            = NewUser("user-key")
	benchmarkSimpleUserJSON        = []byte(`{"key":"user-key"}`)
	benchmarkUserWithAllAttributes = NewUserBuilder("user-key").
					Secondary("s").
					IP("i").
					Country("c").
					Email("e").
					FirstName("f").
					LastName("l").
					Avatar("a").
					Name("n").
					Anonymous(true).
					Custom("attr", ldvalue.String("value")).
					Build()
	benchmarkUserWithAllAttributesJSON = []byte(benchmarkUserWithAllAttributes.String())

	benchmarkJSONResult []byte
)

func BenchmarkUserSerializationWithKeyOnly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkJSONResult, _ = json.Marshal(benchmarkUserWithAllAttributes)
	}
}

func BenchmarkUserSerializationWithAllAttributes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkJSONResult, _ = json.Marshal(benchmarkSimpleUser)
	}
}

func BenchmarkUserDeserializationWithKeyOnly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(benchmarkSimpleUserJSON, &benchmarkUserResult)
	}
}

func BenchmarkUserDeserializationWithAllAttributes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(benchmarkUserWithAllAttributesJSON, &benchmarkUserResult)
	}
}
