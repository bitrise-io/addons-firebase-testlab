package ldvalue

import (
	"encoding/json"
	"testing"
)

func BenchmarkSerializeNull(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkJSONResult, _ = json.Marshal(benchmarkSerializeNullValue)
	}
}

func BenchmarkSerializeBool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkJSONResult, _ = json.Marshal(benchmarkSerializeBoolValue)
	}
}

func BenchmarkSerializeInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkJSONResult, _ = json.Marshal(benchmarkSerializeIntValue)
	}
}

func BenchmarkSerializeFloat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkJSONResult, _ = json.Marshal(benchmarkSerializeFloatValue)
	}
}

func BenchmarkSerializeString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkJSONResult, _ = json.Marshal(benchmarkSerializeStringValue)
	}
}

func BenchmarkSerializeArray(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkJSONResult, _ = json.Marshal(benchmarkSerializeArrayValue)
	}
}

func BenchmarkSerializeObject(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkJSONResult, _ = json.Marshal(benchmarkSerializeObjectValue)
	}
}
