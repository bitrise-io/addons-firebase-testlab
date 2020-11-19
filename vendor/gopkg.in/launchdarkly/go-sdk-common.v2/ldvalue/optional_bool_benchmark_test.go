package ldvalue

import "testing"

func BenchmarkNewOptionalBoolNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkOptBoolResult = NewOptionalBool(benchmarkBoolValue)
	}
}

func BenchmarkNewOptionalBoolFromPointerNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkOptBoolResult = NewOptionalBoolFromPointer(benchmarkBoolPointer)
	}
}

func BenchmarkOptionalBoolValueNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkBoolResult = benchmarkOptBoolWithValue.BoolValue()
	}
}

func BenchmarkOptionalBoolGetNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkBoolResult, _ = benchmarkOptBoolWithValue.Get()
	}
}

func BenchmarkOptionalBoolAsValueNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkValueResult = benchmarkOptBoolWithValue.AsValue()
	}
}
