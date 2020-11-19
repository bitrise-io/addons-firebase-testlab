package ldmodel

import (
	"encoding/json"
	"testing"
)

var (
	benchmarkBytesResult   []byte
	benchmarkErrorResult   error
	benchmarkFlagResult    FeatureFlag
	benchmarkSegmentResult Segment
)

func BenchmarkMarshalFlag(b *testing.B) {
	b.Run("all properties", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			benchmarkBytesResult, benchmarkErrorResult =
				jsonDataModelSerialization{}.MarshalFeatureFlag(flagWithAllProperties)
		}
	})

	b.Run("minimal properties", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			benchmarkBytesResult, benchmarkErrorResult =
				jsonDataModelSerialization{}.MarshalFeatureFlag(flagWithMinimalProperties)
		}
	})
}

func BenchmarkMarshalSegment(b *testing.B) {
	b.Run("all properties", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			benchmarkBytesResult, benchmarkErrorResult =
				jsonDataModelSerialization{}.MarshalSegment(segmentWithAllProperties)
		}
	})

	b.Run("minimal properties", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			benchmarkBytesResult, benchmarkErrorResult =
				jsonDataModelSerialization{}.MarshalSegment(segmentWithMinimalProperties)
		}
	})
}

func BenchmarkUnmarshalFlag(b *testing.B) {
	b.Run("all properties", func(b *testing.B) {
		bytes, _ := json.Marshal(flagWithAllPropertiesJSON)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchmarkFlagResult, benchmarkErrorResult =
				jsonDataModelSerialization{}.UnmarshalFeatureFlag(bytes)
		}
	})

	b.Run("minimal properties", func(b *testing.B) {
		bytes, _ := json.Marshal(flagWithMinimalPropertiesJSON)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchmarkFlagResult, benchmarkErrorResult =
				jsonDataModelSerialization{}.UnmarshalFeatureFlag(bytes)
		}
	})
}

func BenchmarkUnmarshalSegment(b *testing.B) {
	b.Run("all properties", func(b *testing.B) {
		bytes, _ := json.Marshal(segmentWithAllPropertiesJSON)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchmarkSegmentResult, benchmarkErrorResult =
				jsonDataModelSerialization{}.UnmarshalSegment(bytes)
		}
	})

	b.Run("minimal properties", func(b *testing.B) {
		bytes, _ := json.Marshal(segmentWithMinimalPropertiesJSON)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			benchmarkSegmentResult, benchmarkErrorResult =
				jsonDataModelSerialization{}.UnmarshalSegment(bytes)
		}
	})
}
