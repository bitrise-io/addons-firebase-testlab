package ldvalue

import "testing"

func BenchmarkNewOptionalStringNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkOptStringResult = NewOptionalString(benchmarkStringValue)
	}
}

func BenchmarkNewOptionalStringFromPointerNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkOptStringResult = NewOptionalStringFromPointer(benchmarkStringPointer)
	}
}

func BenchmarkOptionalStringValueNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkStringResult = benchmarkOptStringWithValue.StringValue()
	}
}

func BenchmarkOptionalStringGetNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkStringResult, benchmarkBoolResult = benchmarkOptStringWithValue.Get()
	}
}

func BenchmarkOptionalStringAsValueNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkValueResult = benchmarkOptStringWithValue.AsValue()
	}
}
