package ldvalue

import "testing"

func BenchmarkNewOptionalIntNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkOptIntResult = NewOptionalInt(benchmarkIntValue)
	}
}

func BenchmarkNewOptionalIntFromPointerNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkOptIntResult = NewOptionalIntFromPointer(benchmarkIntPointer)
	}
}

func BenchmarkOptionalIntValueNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkIntResult = benchmarkOptIntWithValue.IntValue()
	}
}

func BenchmarkOptionalIntGetNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkIntResult, benchmarkBoolResult = benchmarkOptIntWithValue.Get()
	}
}

func BenchmarkOptionalIntAsValueNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkValueResult = benchmarkOptIntWithValue.AsValue()
	}
}
