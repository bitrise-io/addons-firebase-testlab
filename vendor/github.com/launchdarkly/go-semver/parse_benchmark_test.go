package semver

import "testing"

var (
	// use package-level variables so the compiler won't optimize away benchmark logic
	benchmarkVer Version
	benchmarkErr error
)

// These benchmarks are based on the ones defined in github.com/blang/semver (see:
// https://github.com/blang/semver/blob/master/semver_test.go), not including benchmarks for
// any non-standard parsing modes.

func BenchmarkParseSimple(b *testing.B) {
	const VERSION = "0.0.1"
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchmarkVer, benchmarkErr = Parse(VERSION)
		if benchmarkErr != nil {
			b.Fatal(benchmarkErr)
		}
	}
}

func BenchmarkParseComplex(b *testing.B) {
	const VERSION = "0.0.1-alpha.preview+123.456"
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchmarkVer, benchmarkErr = Parse(VERSION)
		if benchmarkErr != nil {
			b.Fatal(benchmarkErr)
		}
	}
}

func BenchmarkParseAverage(b *testing.B) {
	l := len(benchmarkFormatTests)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchmarkVer, benchmarkErr = Parse(benchmarkFormatTests[n%l])
		if benchmarkErr != nil {
			b.Fatal(benchmarkErr)
		}
	}
}

var benchmarkFormatTests = []string{
	"1.2.3",
	"0.0.1",
	"0.0.1-alpha.preview+123.456",
	"1.2.3-alpha.1+123.456",
	"1.2.3-alpha.1",
	"1.2.3+123.456",
	"1.2.3-alpha.b-eta+123.b-uild",
	"1.2.3+123.b-uild",
	"1.2.3-alpha.b-eta",
}
