package jsonstream

import (
	"encoding/json"
	"testing"
)

var (
	benchmarkBytesResult []byte
	benchmarkErrResult   error
)

type benchmarkJSONStruct struct {
	Field1 string
	Field2 []int
}

// The benchmarks here that test our code have corresponding "JSONMarshal" benchmarks that convert an
// exactly identical data structure to JSON via json.Marshal, so we can compare them for speed and
// memory usage.

func BenchmarkWriteNumbers(b *testing.B) {
	var data1 []int
	var data2 []float64
	for i := 0; i < 50; i++ {
		data1 = append(data1, i*10)
		data2 = append(data2, float64(i)*10.5)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf JSONBuffer
		buf.BeginArray()
		for _, n := range data1 {
			buf.WriteInt(n)
		}
		for _, n := range data2 {
			buf.WriteFloat64(n)
		}
		buf.EndArray()
		benchmarkBytesResult, benchmarkErrResult = buf.Get()
	}
}

func BenchmarkWriteNumbersJSONMarshal(b *testing.B) {
	var data1 []int
	var data2 []float64
	for i := 0; i < 50; i++ {
		data1 = append(data1, i*10)
		data2 = append(data2, float64(i)*10.5)
	}
	data := []interface{}{data1, data2}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkBytesResult, benchmarkErrResult = json.Marshal(data)
	}
}

func BenchmarkStruct(b *testing.B) {
	data := makeStruct()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf JSONBuffer
		buf.BeginObject()
		buf.WriteName("Field1")
		buf.WriteString(data.Field1)
		buf.WriteName("Field2")
		buf.BeginArray()
		for _, n := range data.Field2 {
			buf.WriteInt(n)
		}
		buf.EndArray()
		buf.EndObject()
		benchmarkBytesResult, benchmarkErrResult = buf.Get()
	}
}

func BenchmarkStructJSONMarshal(b *testing.B) {
	data := makeStruct()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkBytesResult, benchmarkErrResult = json.Marshal(data)
	}
}

func makeStruct() benchmarkJSONStruct {
	ret := benchmarkJSONStruct{Field1: "I am a string"}
	for i := 0; i < 10; i++ {
		ret.Field2 = append(ret.Field2, i)
	}
	return ret
}

func BenchmarkStringWithEscaping(b *testing.B) {
	data := makeStringWithEscaping()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf JSONBuffer
		buf.WriteString(data)
		benchmarkBytesResult, benchmarkErrResult = buf.Get()
	}
}

func BenchmarkStringWithEscapingJSONMarshal(b *testing.B) {
	data := makeStringWithEscaping()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkBytesResult, benchmarkErrResult = json.Marshal(data)
	}
}

func makeStringWithEscaping() string {
	return "I'm a string\n\tI want to say \"hello\"\nThat's all\f"
}

func BenchmarkLargeData(b *testing.B) {
	data := makeLargeData()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf JSONBuffer
		buf.BeginArray()
		for _, obj := range data {
			buf.BeginObject()
			for name, value := range obj {
				buf.WriteName(name)
				buf.WriteInt(value)
			}
			buf.EndObject()
		}
		buf.EndArray()
		benchmarkBytesResult, benchmarkErrResult = buf.Get()
	}
}

func BenchmarkLargeDataJSONMarshal(b *testing.B) {
	data := makeLargeData()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkBytesResult, benchmarkErrResult = json.Marshal(data)
	}
}

func makeLargeData() []map[string]int {
	// produces an array with 100 repetitions of {"a":0,"b":1,etc.}
	names := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var data []map[string]int
	n := 0

	for j := 0; j < 100; j++ {
		m := make(map[string]int)
		for k := 0; k < len(names); k++ {
			m[names[k:k+1]] = n
			n++
		}
		data = append(data, m)
	}

	return data
}
