package ldvalue

import (
	"encoding/json"
	"reflect"

	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream"
)

// OptionalBool represents a bool that may or may not have a value. This is similar to using a
// bool pointer to distinguish between a false value and nil, but it is safer because it does not
// expose a pointer to any mutable value.
//
// To create an instance with a bool value, use NewOptionalBool. There is no corresponding method
// for creating an instance with no value; simply use the empty literal OptionalBool{}.
//
//     ob1 := NewOptionalBool(1)
//     ob2 := NewOptionalBool(false) // this has a value which is false
//     ob3 := OptionalBool{}         // this does not have a value
//
// This can also be used as a convenient way to construct a bool pointer within an expression.
// For instance, this example causes myIntPointer to point to the bool value true:
//
//     var myBoolPointer *int = NewOptionalBool(true).AsPointer()
//
// This type is used in the Anonymous property of lduser.User, and for other similar fields in
// the LaunchDarkly Go SDK where a bool value may or may not be defined.
type OptionalBool struct {
	value    bool
	hasValue bool
}

// NewOptionalBool constructs an OptionalBool that has a bool value.
//
// There is no corresponding method for creating an OptionalBool with no value; simply use the
// empty literal OptionalBool{}.
func NewOptionalBool(value bool) OptionalBool {
	return OptionalBool{value: value, hasValue: true}
}

// NewOptionalBoolFromPointer constructs an OptionalBool from a bool pointer. If the pointer is
// non-nil, then the OptionalBool copies its value; otherwise the OptionalBool has no value.
func NewOptionalBoolFromPointer(valuePointer *bool) OptionalBool {
	if valuePointer == nil {
		return OptionalBool{hasValue: false}
	}
	return OptionalBool{value: *valuePointer, hasValue: true}
}

// IsDefined returns true if the OptionalBool contains a bool value, or false if it has no value.
func (o OptionalBool) IsDefined() bool {
	return o.hasValue
}

// BoolValue returns the OptionalBool's value, or false if it has no value.
func (o OptionalBool) BoolValue() bool {
	return o.value
}

// Get is a combination of BoolValue and IsDefined. If the OptionalBool contains a bool value, it
// returns that value and true; otherwise it returns false and false.
func (o OptionalBool) Get() (bool, bool) {
	return o.value, o.hasValue
}

// OrElse returns the OptionalBool's value if it has one, or else the specified fallback value.
func (o OptionalBool) OrElse(valueIfEmpty bool) bool {
	if o.hasValue {
		return o.value
	}
	return valueIfEmpty
}

// AsPointer returns the OptionalBool's value as a bool pointer if it has a value, or nil
// otherwise.
//
// The bool value, if any, is copied rather than returning to a pointer to the internal field.
func (o OptionalBool) AsPointer() *bool {
	if o.hasValue {
		v := o.value
		return &v
	}
	return nil
}

// AsValue converts the OptionalBool to a Value, which is either Null() or a boolean value.
func (o OptionalBool) AsValue() Value {
	if o.hasValue {
		return Bool(o.value)
	}
	return Null()
}

// String is a debugging convenience method that returns a description of the OptionalBool. This
// is either "true", "false, or "[none]" if it has no value.
func (o OptionalBool) String() string {
	if o.hasValue {
		if o.value {
			return trueString
		}
		return falseString
	}
	return noneDescription
}

// MarshalJSON converts the OptionalBool to its JSON representation.
//
// The output will be either a JSON boolean or null. Note that the "omitempty" tag for a struct
// field will not cause an empty OptionalBool field to be omitted; it will be output as null.
// If you want to completely omit a JSON property when there is no value, it must be a bool
// pointer instead of an OptionalBool; use the AsPointer() method to get a pointer.
func (o OptionalBool) MarshalJSON() ([]byte, error) {
	return o.AsValue().MarshalJSON()
}

// UnmarshalJSON parses an OptionalBool from JSON.
//
// The input must be either a JSON number that is a boolean or null.
func (o *OptionalBool) UnmarshalJSON(data []byte) error {
	var v Value
	if err := v.UnmarshalJSON(data); err != nil {
		return err // COVERAGE: should not be possible, parser normally doesn't pass malformed content to UnmarshalJSON
	}
	switch {
	case v.IsNull():
		*o = OptionalBool{}
	case v.IsBool():
		*o = NewOptionalBool(v.BoolValue())
	default:
		*o = OptionalBool{}
		return &json.UnmarshalTypeError{Value: string(data), Type: reflect.TypeOf(o)}
	}
	return nil
}

// WriteToJSONBuffer provides JSON serialization for OptionalBool with the jsonstream API.
//
// The JSON output format is identical to what is produced by json.Marshal, but this implementation is
// more efficient when building output with JSONBuffer. See the jsonstream package for more details.
func (o OptionalBool) WriteToJSONBuffer(j *jsonstream.JSONBuffer) {
	o.AsValue().WriteToJSONBuffer(j)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (o OptionalBool) MarshalText() ([]byte, error) {
	if o.hasValue {
		if o.value {
			return []byte(trueString), nil
		}
		return []byte(falseString), nil
	}
	return []byte(""), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// This allows OptionalBool to be used with packages that can parse text content, such as gcfg.
func (o *OptionalBool) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*o = OptionalBool{}
		return nil
	}
	return o.UnmarshalJSON(text)
}
