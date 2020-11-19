package ldvalue

import (
	"encoding/json"
	"testing"

	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream"

	"github.com/stretchr/testify/assert"
)

func TestEmptyOptionalInt(t *testing.T) {
	o := OptionalInt{}
	assert.False(t, o.IsDefined())
	assert.Equal(t, 0, o.IntValue())

	n, ok := o.Get()
	assert.Equal(t, 0, n)
	assert.False(t, ok)

	assert.Equal(t, 2, o.OrElse(2))
	assert.Nil(t, o.AsPointer())
	assert.Equal(t, Null(), o.AsValue())
	assert.True(t, o == o)
}

func TestOptionalIntWithValue(t *testing.T) {
	o := NewOptionalInt(3)
	assert.True(t, o.IsDefined())
	assert.Equal(t, 3, o.IntValue())

	n, ok := o.Get()
	assert.Equal(t, 3, n)
	assert.True(t, ok)

	assert.Equal(t, 3, o.OrElse(2))
	assert.NotNil(t, o.AsPointer())
	assert.Equal(t, 3, *o.AsPointer())
	assert.Equal(t, Int(3), o.AsValue())
	assert.True(t, o == o)
	assert.False(t, o == OptionalInt{})
}

func TestOptionalIntFromNilPointer(t *testing.T) {
	o := NewOptionalIntFromPointer(nil)
	assert.True(t, o == OptionalInt{})
}

func TestOptionalIntFromNonNilPointer(t *testing.T) {
	v := 3
	p := &v
	o := NewOptionalIntFromPointer(p)
	assert.True(t, o == NewOptionalInt(3))

	assert.Equal(t, 3, *o.AsPointer())
	assert.False(t, p == o.AsPointer()) // should not be the same pointer, just the same underlying value
}

func TestOptionalIntAsStringer(t *testing.T) {
	assert.Equal(t, "[none]", OptionalInt{}.String())
	assert.Equal(t, "3", NewOptionalInt(3).String())
}

func TestOptionalIntJSONMarshalling(t *testing.T) {
	bytes, err := json.Marshal(OptionalInt{})
	assert.NoError(t, err)
	assert.Equal(t, nullAsJSON, string(bytes))

	bytes, err = json.Marshal(NewOptionalInt(3))
	assert.NoError(t, err)
	assert.Equal(t, `3`, string(bytes))

	swos := structWithOptionalInts{N1: NewOptionalInt(3)}
	bytes, err = json.Marshal(swos)
	assert.NoError(t, err)
	assert.Equal(t, `{"n1":3,"n2":null,"n3":null}`, string(bytes))

	var j jsonstream.JSONBuffer
	j.SetSeparator([]byte(","))
	NewOptionalInt(3).WriteToJSONBuffer(&j)
	OptionalInt{}.WriteToJSONBuffer(&j)
	bytes, err = j.Get()
	assert.NoError(t, err)
	assert.Equal(t, `3,null`, string(bytes))
}

func TestOptionalIntJSONUnmarshalling(t *testing.T) {
	var o OptionalInt
	err := json.Unmarshal([]byte(nullAsJSON), &o)
	assert.NoError(t, err)
	assert.False(t, o.IsDefined())

	err = json.Unmarshal([]byte(`3`), &o)
	assert.NoError(t, err)
	assert.Equal(t, NewOptionalInt(3), o)

	err = json.Unmarshal([]byte(`true`), &o)
	assert.Error(t, err)
	assert.IsType(t, &json.UnmarshalTypeError{}, err)

	err = json.Unmarshal([]byte(`x`), &o)
	assert.Error(t, err)
	assert.IsType(t, &json.SyntaxError{}, err)

	var swos structWithOptionalInts
	err = json.Unmarshal([]byte(`{"n1":3,"n3":null}`), &swos)
	assert.NoError(t, err)
	assert.Equal(t, NewOptionalInt(3), swos.N1)
	assert.Equal(t, OptionalInt{}, swos.N2)
	assert.Equal(t, OptionalInt{}, swos.N3)
}

type structWithOptionalInts struct {
	N1 OptionalInt `json:"n1"`
	N2 OptionalInt `json:"n2"`
	N3 OptionalInt `json:"n3"`
}

func TestOptionalIntTextMarshalling(t *testing.T) {
	b, e := NewOptionalInt(3).MarshalText()
	assert.NoError(t, e)
	assert.Equal(t, []byte("3"), b)

	b, e = OptionalInt{}.MarshalText()
	assert.NoError(t, e)
	assert.Len(t, b, 0)
}

func TestOptionalIntTextUnmarshalling(t *testing.T) {
	var o1 OptionalInt
	assert.NoError(t, o1.UnmarshalText([]byte("3")))
	assert.Equal(t, NewOptionalInt(3), o1)

	var o2 OptionalInt
	assert.NoError(t, o2.UnmarshalText([]byte("")))
	assert.Equal(t, OptionalInt{}, o2)

	var o3 OptionalInt
	assert.NoError(t, o3.UnmarshalText(nil))
	assert.Equal(t, OptionalInt{}, o3)

	var o4 OptionalInt
	assert.Error(t, o4.UnmarshalText([]byte("x")))
	assert.Equal(t, OptionalInt{}, o4)
}
