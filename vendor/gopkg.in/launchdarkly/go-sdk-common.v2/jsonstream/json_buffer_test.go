package jsonstream

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type valueTest struct {
	name     string
	encoding string
	action   func(j *JSONBuffer)
}

var allValueTests = []valueTest{
	{"null", "null", func(j *JSONBuffer) { j.WriteNull() }},
	{"bool true", "true", func(j *JSONBuffer) { j.WriteBool(true) }},
	{"bool true", "false", func(j *JSONBuffer) { j.WriteBool(false) }},
	{"int", "3", func(j *JSONBuffer) { j.WriteInt(3) }},
	{"zero int", "0", func(j *JSONBuffer) { j.WriteInt(0) }},
	{"negative int", "-3", func(j *JSONBuffer) { j.WriteInt(-3) }},
	{"large int", "123456", func(j *JSONBuffer) { j.WriteInt(123456) }},
	{"negative large int", "-123456", func(j *JSONBuffer) { j.WriteInt(-123456) }},
	{"uint64", "3", func(j *JSONBuffer) { j.WriteUint64(3) }},
	{"zero uint64", "0", func(j *JSONBuffer) { j.WriteUint64(0) }},
	{"large uint64", "12345678901234567890", func(j *JSONBuffer) { j.WriteUint64(12345678901234567890) }},
	{"float", "3.5", func(j *JSONBuffer) { j.WriteFloat64(3.5) }},
	{"zero float", "0", func(j *JSONBuffer) { j.WriteFloat64(0) }},
	{"integer as float", "3", func(j *JSONBuffer) { j.WriteFloat64(float64(3)) }},
	{"negative float", "-3.5", func(j *JSONBuffer) { j.WriteFloat64(-3.5) }},
	{"string", `"abc"`, func(j *JSONBuffer) { j.WriteString("abc") }},
	{"empty string", `""`, func(j *JSONBuffer) { j.WriteString("") }},
	{"array", `[]`, func(j *JSONBuffer) {
		j.BeginArray()
		j.EndArray()
	}},
	{"object", `{}`, func(j *JSONBuffer) {
		j.BeginObject()
		j.EndObject()
	}},
	{"raw", `{"a":1}`, func(j *JSONBuffer) { j.WriteRaw(json.RawMessage(`{"a":1}`)) }},
}

type escapeTest struct {
	decoded string
	encoded string
}

var basicEscapeTests = []escapeTest{
	{"\b", `\b`}, {"\t", `\t`}, {"\n", `\n`}, {"\f", `\f`}, {"\r", `\r`}, {`"`, `\"`}, {`\`, `\\`},
	{"\x05", `\u0005`}, {"\x1c", `\u001c`},
	{"🦜🦄😂🧶😻 yes", "🦜🦄😂🧶😻 yes"}, // unescaped multi-byte characters are allowed
}

func testEncoding(t *testing.T, name string, expected string, action func(*JSONBuffer)) {
	t.Run(name, func(t *testing.T) {
		j := NewJSONBuffer()
		action(j)
		bytes, err := j.Get()
		if err == nil {
			assert.Equal(t, expected, string(bytes))
		} else {
			assert.NoError(t, err, "data written to buffer so far: %s", j.GetPartial())
		}
	})
}

func makeStringEscapingTests(forParser bool) []escapeTest {
	allBasic := basicEscapeTests
	if forParser {
		// These escapes are not used when writing, but may be encountered when parsing
		allBasic = append(allBasic, escapeTest{"/", `\/`})
		allBasic = append(allBasic, escapeTest{"も", `\3082`})
	}
	// This creates various permutations to ensure that string escaping is handled correctly regardless of
	// whether the escape sequence is at the beginning, the end, next to another escape sequence, etc.
	var ret []escapeTest
	for _, test := range allBasic {
		ret = append(ret, test)
		for _, f := range []string{"%sabcd", "abcd%s", "ab%scd"} {
			ret = append(ret, escapeTest{decoded: fmt.Sprintf(f, test.decoded),
				encoded: fmt.Sprintf(f, test.encoded)})
		}
		for _, test2 := range allBasic {
			for _, f := range []string{"%s%sabcd", "ab%s%scd", "a%sbc%sd", "abcd%s%s"} {
				ret = append(ret, escapeTest{decoded: fmt.Sprintf(f, test.decoded, test2.decoded),
					encoded: fmt.Sprintf(f, test.encoded, test2.encoded)})
			}
		}
	}
	return ret
}

func TestWriteSimpleValues(t *testing.T) {
	for _, test := range allValueTests {
		testEncoding(t, test.name, test.encoding, test.action)
	}
}

func TestWriteStringFormatting(t *testing.T) {
	for _, test := range makeStringEscapingTests(false) {
		testEncoding(t, test.encoded, `"`+test.encoded+`"`, func(j *JSONBuffer) {
			j.WriteString(test.decoded)
		})
	}
}

func TestWriteArray(t *testing.T) {
	testEncoding(t, "empty", "[]", func(j *JSONBuffer) {
		j.BeginArray()
		j.EndArray()
	})

	for _, test := range allValueTests {
		ve := test.encoding
		t.Run("value = "+test.name, func(t *testing.T) {
			testEncoding(t, "single", `[`+ve+`]`, func(j *JSONBuffer) {
				j.BeginArray()
				test.action(j)
				j.EndArray()
			})

			testEncoding(t, "multiple", `[`+ve+`,`+ve+`]`, func(j *JSONBuffer) {
				j.BeginArray()
				test.action(j)
				test.action(j)
				j.EndArray()
			})

			testEncoding(t, "nested", `[`+ve+`,[`+ve+`,`+ve+`],`+ve+`]`, func(j *JSONBuffer) {
				j.BeginArray()
				test.action(j)
				j.BeginArray()
				test.action(j)
				test.action(j)
				j.EndArray()
				test.action(j)
				j.EndArray()
			})
		})
	}
}

func TestWriteObject(t *testing.T) {
	testEncoding(t, "empty", "{}", func(j *JSONBuffer) {
		j.BeginObject()
		j.EndObject()
	})

	for _, test := range allValueTests {
		ve := test.encoding
		t.Run("value = "+test.name, func(t *testing.T) {
			testEncoding(t, "single", `{"a":`+ve+`}`, func(j *JSONBuffer) {
				j.BeginObject()
				j.WriteName("a")
				test.action(j)
				j.EndObject()
			})

			testEncoding(t, "multiple", `{"a":`+ve+`,"b":`+ve+`}`, func(j *JSONBuffer) {
				j.BeginObject()
				j.WriteName("a")
				test.action(j)
				j.WriteName("b")
				test.action(j)
				j.EndObject()
			})

			testEncoding(t, "nested", `{"a":{"b":`+ve+`,"c":`+ve+`},"d":`+ve+`}`, func(j *JSONBuffer) {
				j.BeginObject()
				j.WriteName("a")
				j.BeginObject()
				j.WriteName("b")
				test.action(j)
				j.WriteName("c")
				test.action(j)
				j.EndObject()
				j.WriteName("d")
				test.action(j)
				j.EndObject()
			})
		})
	}
}

func TestWriteMultipleValues(t *testing.T) {
	testEncoding(t, "default separator", "1\n2\n3", func(j *JSONBuffer) {
		j.WriteInt(1)
		j.WriteInt(2)
		j.WriteInt(3)
	})

	testEncoding(t, "nil is same as default separator", "1\n2\n3", func(j *JSONBuffer) {
		j.SetSeparator(nil)
		j.WriteInt(1)
		j.WriteInt(2)
		j.WriteInt(3)
	})

	testEncoding(t, "custom separator", "1! 2! 3", func(j *JSONBuffer) {
		j.SetSeparator([]byte("! "))
		j.WriteInt(1)
		j.WriteInt(2)
		j.WriteInt(3)
	})
}

func TestWriteDeeplyNestedStructures(t *testing.T) {
	expected := ""
	var j JSONBuffer
	for i := 0; i < 100; i++ {
		j.BeginObject()
		j.WriteName("x")
		j.BeginArray()
		expected += `{"x":[`
	}
	for i := 0; i < 100; i++ {
		j.EndArray()
		j.EndObject()
		expected += "]}"
	}
	bytes, err := j.Get()
	assert.NoError(t, err)
	assert.Equal(t, expected, string(bytes))
}

func TestUsageErrors(t *testing.T) {
	shouldFail := func(name string, expectedPartialOutput string, action func(j *JSONBuffer)) {
		t.Run(name, func(t *testing.T) {
			j := NewJSONBuffer()
			action(j)
			bytes, err := j.Get()
			assert.Error(t, err)
			assert.Nil(t, bytes, "data written to buffer: %s", bytes)

			assert.Equal(t, err, j.GetError())
			assert.Equal(t, expectedPartialOutput, string(j.GetPartial()))
		})
	}

	shouldFail("BeginArray not closed", `[`, func(j *JSONBuffer) { j.BeginArray() })
	shouldFail("BeginObject not closed", `{`, func(j *JSONBuffer) { j.BeginObject() })

	shouldFail("EndArray before anything", ``, func(j *JSONBuffer) { j.EndArray() })
	shouldFail("EndArray in object before name", `{`, func(j *JSONBuffer) {
		j.BeginObject()
		j.EndArray()
	})
	shouldFail("EndArray in object after name", `{"name":`, func(j *JSONBuffer) {
		j.BeginObject()
		j.WriteName("name")
		j.EndArray()
	})
	shouldFail("EndArray in object after value", `{"name":1`, func(j *JSONBuffer) {
		j.BeginObject()
		j.WriteName("name")
		j.WriteInt(1)
		j.EndArray()
	})

	shouldFail("EndObject before anything", ``, func(j *JSONBuffer) { j.EndObject() })
	shouldFail("EndObject when value is expected", `{"name":`, func(j *JSONBuffer) {
		j.BeginObject()
		j.WriteName("name")
		j.EndObject()
	})
	shouldFail("EndObject in array before value", `[`, func(j *JSONBuffer) {
		j.BeginArray()
		j.EndObject()
	})
	shouldFail("EndObject in array after value", `[1`, func(j *JSONBuffer) {
		j.BeginArray()
		j.WriteInt(1)
		j.EndObject()
	})

	shouldFail("WriteName when value is expected", `[`, func(j *JSONBuffer) {
		j.BeginArray()
		j.WriteName("name")
	})
	shouldFail("WriteName after an error has already occurred", `{`, func(j *JSONBuffer) {
		j.BeginObject()
		j.EndArray()
		j.WriteName("name")
	})

	for _, test := range allValueTests {
		shouldFail(test.name+" when first property name is expected", `{`, func(j *JSONBuffer) {
			j.BeginObject()
			test.action(j)
			j.EndObject()
		})

		shouldFail(test.name+" when subsequent property name is expected", `{"x":true`, func(j *JSONBuffer) {
			j.BeginObject()
			j.WriteName("x")
			j.WriteBool(true)
			test.action(j)
			j.EndObject()
		})

		shouldFail(test.name+" after an error has already occurred", `1`, func(j *JSONBuffer) {
			j.WriteInt(1)
			j.EndArray() // puts buffer in a failed state
			test.action(j)
		})
	}
}

func TestGrowDoesNotAffectOutput(t *testing.T) {
	j := NewJSONBuffer()
	j.WriteInt(1)
	j.Grow(100)
	j.WriteInt(2)
	bytes, err := j.Get()
	assert.NoError(t, err)
	assert.Equal(t, "1\n2", string(bytes))
}

func TestJSONBufferCanBeLocallyAllocated(t *testing.T) {
	var j JSONBuffer
	j.WriteBool(true)
	bytes, err := j.Get()
	assert.NoError(t, err)
	assert.Equal(t, "true", string(bytes))
}

func TestStreamingJSONBuffer(t *testing.T) {
	t.Run("data is flushed incrementally", func(t *testing.T) {
		var target bytes.Buffer
		j := NewStreamingJSONBuffer(&target, 20)
		j.SetSeparator([]byte(","))
		j.WriteInt(123456789012345)
		assert.Len(t, target.Bytes(), 0)
		j.WriteInt(123456)
		assert.Equal(t, `123456789012345,123456`, string(target.Bytes()))
		j.WriteInt(12345678)
		assert.Equal(t, `123456789012345,123456`, string(target.Bytes()))
		j.Flush()
		assert.Equal(t, `123456789012345,123456,12345678`, string(target.Bytes()))
	})

	t.Run("writer error prevents subsequent writes", func(t *testing.T) {
		e := errors.New("sorry")
		var w testWriter
		j := NewStreamingJSONBuffer(&w, 20)
		j.SetSeparator([]byte(","))
		j.WriteInt(12345)
		j.Flush()
		j.WriteInt(67890)
		w.fakeError = e
		j.Flush()
		j.WriteInt(22222)
		j.WriteInt(33333)
		j.Flush()
		assert.Equal(t, e, j.GetError())
		assert.Equal(t, "12345", string(w.target.Bytes()))
	})
}

type testWriter struct {
	target    bytes.Buffer
	fakeError error
}

func (w *testWriter) Write(data []byte) (int, error) {
	if w.fakeError != nil {
		return 0, w.fakeError
	}
	return w.target.Write(data)
}
