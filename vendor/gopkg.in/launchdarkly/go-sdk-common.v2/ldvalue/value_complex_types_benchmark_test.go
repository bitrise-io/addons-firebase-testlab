package ldvalue

import "testing"

func BenchmarkArrayBuild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkValueResult = ArrayBuild().Add(Int(1)).Add(Int(2)).Add(Int(3)).Build()
	}
}

func BenchmarkArrayOf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkValueResult = ArrayOf(Int(1), Int(2), Int(3))
	}
}

func BenchmarkObjectBuild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkValueResult = ObjectBuild().Set("a", Int(1)).Set("b", Int(2)).Set("c", Int(3)).Build()
	}
}
