package ldvalue

import "testing"

func BenchmarkNullValueNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkValueResult = Null()
	}
}

func BenchmarkBoolValueNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkValueResult = Bool(true)
		benchmarkBoolResult = benchmarkValueResult.BoolValue()
	}
}

func BenchmarkIntValueNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkValueResult = Int(1)
		benchmarkIntResult = benchmarkValueResult.IntValue()
	}
}

func BenchmarkFloat64ValueNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkValueResult = Float64(1)
		benchmarkFloat64Result = benchmarkValueResult.Float64Value()
	}
}

func BenchmarkStringValueNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkValueResult = String(benchmarkStringValue)
		benchmarkStringResult = benchmarkValueResult.StringValue()
	}
}
