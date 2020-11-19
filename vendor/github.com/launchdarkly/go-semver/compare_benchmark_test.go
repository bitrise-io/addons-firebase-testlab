package semver

import "testing"

// These benchmarks are based on the ones defined in github.com/blang/semver (see:
// https://github.com/blang/semver/blob/master/semver_test.go), not including benchmarks for
// any non-standard parsing modes or range comparisons.

var (
	// use package-level variables so the compiler won't optimize away benchmark logic
	benchmarkCompareResult int
)

func BenchmarkCompareSimple(b *testing.B) {
	const VERSION = "0.0.1"
	v, _ := Parse(VERSION)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchmarkCompareResult = v.ComparePrecedence(v)
		if benchmarkCompareResult != 0 {
			b.Fail()
		}
	}
}

func BenchmarkCompareComplex(b *testing.B) {
	const VERSION = "0.0.1-alpha.preview+123.456"
	v, _ := Parse(VERSION)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		benchmarkCompareResult = v.ComparePrecedence(v)
		if benchmarkCompareResult != 0 {
			b.Fail()
		}
	}
}

func BenchmarkCompareAverage(b *testing.B) {
	l := len(compareTests)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		test := compareTests[n%l]
		benchmarkCompareResult = test.v1.ComparePrecedence(test.v2)
		if benchmarkCompareResult != test.result {
			b.Fail()
		}
	}
}
