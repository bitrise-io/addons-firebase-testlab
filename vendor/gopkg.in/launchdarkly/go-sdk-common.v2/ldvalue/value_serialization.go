package ldvalue

import (
	"encoding/json"
	"errors"
	"strconv"

	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream"
)

// This file contains methods for converting Value to and from JSON.

// Parse returns a Value parsed from a JSON string, or Null if it cannot be parsed.
//
// This is simply a shortcut for calling json.Unmarshal and disregarding errors. It is meant for
// use in test scenarios where malformed data is not a concern.
func Parse(jsonData []byte) Value {
	var v Value
	if err := json.Unmarshal(jsonData, &v); err != nil {
		return Null()
	}
	return v
}

// JSONString returns the JSON representation of the value.
func (v Value) JSONString() string {
	// The following is somewhat redundant with json.Marshal, but it avoids the overhead of
	// converting between byte arrays and strings.
	switch v.valueType {
	case NullType:
		return nullAsJSON
	case BoolType:
		if v.boolValue {
			return trueString
		}
		return falseString
	case NumberType:
		if v.IsInt() {
			return strconv.Itoa(int(v.numberValue))
		}
		return strconv.FormatFloat(v.numberValue, 'f', -1, 64)
	}
	// For all other types, we rely on our custom marshaller.
	bytes, _ := json.Marshal(v)
	// It shouldn't be possible for marshalling to fail, because Value can only contain
	// JSON-compatible types. But if it somehow did fail, bytes will be nil and we'll return
	// an empty tring.
	return string(bytes)
}

// MarshalJSON converts the Value to its JSON representation.
//
// Note that the "omitempty" tag for a struct field will not cause an empty Value field to be
// omitted; it will be output as null. If you want to completely omit a JSON property when there
// is no value, it must be a pointer; use AsPointer().
func (v Value) MarshalJSON() ([]byte, error) {
	switch v.valueType {
	case NullType:
		return []byte(nullAsJSON), nil
	case BoolType:
		if v.boolValue {
			return []byte(trueString), nil
		}
		return []byte(falseString), nil
	case NumberType:
		if v.IsInt() {
			return []byte(strconv.Itoa(int(v.numberValue))), nil
		}
		return []byte(strconv.FormatFloat(v.numberValue, 'f', -1, 64)), nil
	case StringType:
		return json.Marshal(v.stringValue)
	case ArrayType:
		if v.immutableArrayValue == nil {
			return json.Marshal([]Value{})
		}
		return json.Marshal(v.immutableArrayValue)
	case ObjectType:
		if v.immutableObjectValue == nil {
			return json.Marshal(map[string]Value{})
		}
		return json.Marshal(v.immutableObjectValue)
	case RawType:
		return []byte(v.stringValue), nil
	}
	return nil, errors.New("unknown data type") // should not be possible
}

// UnmarshalJSON parses a Value from JSON.
func (v *Value) UnmarshalJSON(data []byte) error { //nolint:funlen // yes, we know it's a long function
	if len(data) == 0 { // COVERAGE: should not be possible, parser doesn't pass empty slices to UnmarshalJSON
		return errors.New("cannot parse empty data")
	}
	firstCh := data[0]
	switch firstCh {
	case 'n':
		// Note that since Go 1.5, comparing a string to string([]byte) is optimized so it
		// does not actually create a new string from the byte slice.
		if string(data) == "null" {
			*v = Null()
			return nil
		}
	case 't', 'f':
		if string(data) == trueString {
			*v = Bool(true)
			return nil
		}
		if string(data) == falseString {
			*v = Bool(false)
			return nil
		}
	case '"':
		var s string
		e := json.Unmarshal(data, &s)
		if e == nil {
			*v = String(s)
		}
		return e
	case '[':
		var a []Value
		e := json.Unmarshal(data, &a)
		if e == nil {
			if len(a) == 0 {
				a = nil // don't need to retain an empty array
			}
			*v = Value{valueType: ArrayType, immutableArrayValue: a}
		}
		return e
	case '{':
		var o map[string]Value
		e := json.Unmarshal(data, &o)
		if e == nil {
			if len(o) == 0 {
				o = nil // don't need to retain an empty map
			}
			*v = Value{valueType: ObjectType, immutableObjectValue: o}
		}
		return e
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9': // note, JSON does not allow a leading '.'
		var n float64
		e := json.Unmarshal(data, &n)
		if e == nil {
			*v = Value{valueType: NumberType, numberValue: n}
		}
		return e
	}
	return &json.SyntaxError{} // COVERAGE: never happens, parser rejects the token earlier
}

// WriteToJSONBuffer provides JSON serialization for Value with the jsonstream API.
//
// The JSON output format is identical to what is produced by json.Marshal, but this implementation is
// more efficient when building output with JSONBuffer. See the jsonstream package for more details.
func (v Value) WriteToJSONBuffer(j *jsonstream.JSONBuffer) {
	switch v.valueType {
	case NullType:
		j.WriteNull()
	case BoolType:
		j.WriteBool(v.boolValue)
	case NumberType:
		j.WriteFloat64(v.numberValue)
	case StringType:
		j.WriteString(v.stringValue)
	case ArrayType:
		j.BeginArray()
		for _, vv := range v.immutableArrayValue {
			vv.WriteToJSONBuffer(j)
		}
		j.EndArray()
	case ObjectType:
		j.BeginObject()
		for k, vv := range v.immutableObjectValue {
			j.WriteName(k)
			vv.WriteToJSONBuffer(j)
		}
		j.EndObject()
	case RawType:
		j.WriteRaw([]byte(v.stringValue))
	}
}
