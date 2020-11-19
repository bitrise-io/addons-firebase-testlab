package ldvalue

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	helpers "github.com/launchdarkly/go-test-helpers/v2"
)

func TestValueTypes(t *testing.T) {
	assert.Equal(t, nullAsJSON, NullType.String())
	assert.Equal(t, "bool", BoolType.String())
	assert.Equal(t, "number", NumberType.String())
	assert.Equal(t, "string", StringType.String())
	assert.Equal(t, "array", ArrayType.String())
	assert.Equal(t, "object", ObjectType.String())
	assert.Equal(t, "raw", RawType.String())
	assert.Equal(t, "unknown", ValueType(99).String())
}

func TestNullValue(t *testing.T) {
	v := Null()

	assert.Equal(t, NullType, v.Type())
	assert.True(t, v.IsNull())
	assert.False(t, v.IsBool())
	assert.False(t, v.IsNumber())
	assert.False(t, v.IsInt())
	assert.False(t, v.IsString())

	assert.Equal(t, Null(), v)
	assert.Equal(t, Value{}, v)

	// treating a null as a non-null produces empty values
	assert.False(t, v.BoolValue())
	assert.Equal(t, 0, v.IntValue())
	assert.Equal(t, float64(0), v.Float64Value())
	assert.Equal(t, "", v.StringValue())
	assert.Equal(t, OptionalString{}, v.AsOptionalString())
	assert.Equal(t, 0, v.Count())
	assert.Equal(t, Null(), v.GetByIndex(0))
	assert.Equal(t, Null(), v.GetByKey("x"))
}

func TestBoolValue(t *testing.T) {
	tv := Bool(true)

	assert.Equal(t, BoolType, tv.Type())
	assert.True(t, tv.BoolValue())
	assert.False(t, tv.IsNull())
	assert.True(t, tv.IsBool())
	assert.False(t, tv.IsNumber())
	assert.False(t, tv.IsInt())
	assert.False(t, tv.IsString())

	assert.Equal(t, Bool(true), tv)
	assert.NotEqual(t, Bool(false), tv)

	// treating a bool as a non-bool produces empty values
	assert.Equal(t, 0, tv.IntValue())
	assert.Equal(t, float64(0), tv.Float64Value())
	assert.Equal(t, "", tv.StringValue())
	assert.Equal(t, OptionalString{}, tv.AsOptionalString())
	assert.Equal(t, 0, tv.Count())
	assert.Equal(t, Null(), tv.GetByIndex(0))
	assert.Equal(t, Null(), tv.GetByKey("x"))

	fv := Bool(false)

	assert.Equal(t, BoolType, fv.Type())
	assert.False(t, fv.BoolValue())
	assert.False(t, fv.IsNull())
	assert.False(t, fv.IsNumber())
	assert.False(t, fv.IsInt())

	assert.Equal(t, Bool(false), fv)
	assert.NotEqual(t, Bool(true), fv)
}

func TestIntValue(t *testing.T) {
	v := Int(2)

	assert.Equal(t, NumberType, v.Type())
	assert.Equal(t, 2, v.IntValue())
	assert.Equal(t, float64(2), v.Float64Value())
	assert.False(t, v.IsNull())
	assert.False(t, v.IsBool())
	assert.True(t, v.IsNumber())
	assert.True(t, v.IsInt())
	assert.False(t, v.IsString())

	assert.Equal(t, Int(2), v)
	assert.Equal(t, Float64(2), v)
	assert.NotEqual(t, Float64(2.5), v)

	// treating a number as a non-number produces empty values
	assert.False(t, v.BoolValue())
	assert.Equal(t, "", v.StringValue())
	assert.Equal(t, OptionalString{}, v.AsOptionalString())
	assert.Equal(t, 0, v.Count())
	assert.Equal(t, Null(), v.GetByIndex(0))
	assert.Equal(t, Null(), v.GetByKey("x"))
}

func TestFloat64Value(t *testing.T) {
	v := Float64(2.75)

	assert.Equal(t, NumberType, v.Type())
	assert.Equal(t, 2, v.IntValue())
	assert.Equal(t, 2.75, v.Float64Value())
	assert.False(t, v.IsNull())
	assert.False(t, v.IsBool())
	assert.True(t, v.IsNumber())
	assert.False(t, v.IsInt())
	assert.False(t, v.IsString())

	floatButReallyInt := Float64(2.0)
	assert.Equal(t, NumberType, floatButReallyInt.Type())
	assert.Equal(t, 2, floatButReallyInt.IntValue())
	assert.Equal(t, 2.0, floatButReallyInt.Float64Value())
	assert.False(t, floatButReallyInt.IsNull())
	assert.True(t, floatButReallyInt.IsNumber())
	assert.True(t, floatButReallyInt.IsInt())

	assert.Equal(t, Float64(2.75), v)
	assert.NotEqual(t, Float64(2.5), v)

	// treating a number as a non-number produces empty values
	assert.False(t, v.BoolValue())
	assert.Equal(t, "", v.StringValue())
	assert.Equal(t, OptionalString{}, v.AsOptionalString())
	assert.Equal(t, 0, v.Count())
	assert.Equal(t, Null(), v.GetByIndex(0))
	assert.Equal(t, Null(), v.GetByKey("x"))
}

func TestStringValue(t *testing.T) {
	v := String("abc")

	assert.Equal(t, StringType, v.Type())
	assert.Equal(t, "abc", v.StringValue())
	assert.Equal(t, NewOptionalString("abc"), v.AsOptionalString())
	assert.False(t, v.IsNull())
	assert.False(t, v.IsBool())
	assert.False(t, v.IsNumber())
	assert.False(t, v.IsInt())
	assert.True(t, v.IsString())
	assert.Equal(t, v, String("abc"))

	assert.Equal(t, String("abc"), v)
	assert.NotEqual(t, String("def"), v)

	// treating a string as a non-string produces empty values
	assert.False(t, v.BoolValue())
	assert.Equal(t, 0, v.IntValue())
	assert.Equal(t, float64(0), v.Float64Value())
	assert.Equal(t, 0, v.Count())
	assert.Equal(t, Null(), v.GetByIndex(0))
	assert.Equal(t, Null(), v.GetByKey("x"))
}

func TestRawValue(t *testing.T) {
	rawJson := json.RawMessage([]byte("[1]"))
	v := Raw(rawJson)

	assert.Equal(t, RawType, v.Type())
	assert.Equal(t, rawJson, v.AsRaw())
	assert.False(t, v.IsNull())
	assert.False(t, v.IsBool())
	assert.False(t, v.IsNumber())
	assert.False(t, v.IsInt())
	assert.False(t, v.IsString())

	// conversion of other types to Raw is covered in value_serialization_test

	// treating a Raw as a non-Raw produces empty values
	assert.False(t, v.BoolValue())
	assert.Equal(t, 0, v.IntValue())
	assert.Equal(t, float64(0), v.Float64Value())
	assert.Equal(t, "", v.StringValue())
	assert.Equal(t, OptionalString{}, v.AsOptionalString())
	assert.Equal(t, 0, v.Count())
	assert.Equal(t, Null(), v.GetByIndex(0))
	assert.Equal(t, Null(), v.GetByKey("x"))
}

func TestConvertPrimitivesFromArbitraryValue(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Equal(t, Null(), CopyArbitraryValue(nil))
	})
	t.Run("Value", func(t *testing.T) {
		originalValue := Int(1)
		assert.Equal(t, originalValue, CopyArbitraryValue(originalValue))
		assert.Equal(t, originalValue, CopyArbitraryValue(&originalValue))
		assert.Equal(t, Null(), CopyArbitraryValue((*Value)(nil)))
	})
	t.Run("OptionalString", func(t *testing.T) {
		s := NewOptionalString("value")
		sv := String("value")
		assert.Equal(t, sv, CopyArbitraryValue(s))
		assert.Equal(t, sv, CopyArbitraryValue(&s))
		assert.Equal(t, Null(), CopyArbitraryValue(OptionalString{}))
		assert.Equal(t, Null(), CopyArbitraryValue(&OptionalString{}))
		assert.Equal(t, Null(), CopyArbitraryValue((*OptionalString)(nil)))
	})
	t.Run("bool", func(t *testing.T) {
		assert.Equal(t, Bool(true), CopyArbitraryValue(true))
		assert.Equal(t, Bool(false), CopyArbitraryValue(false))
		assert.Equal(t, Bool(true), CopyArbitraryValue(helpers.BoolPtr(true)))
		assert.Equal(t, Bool(false), CopyArbitraryValue(helpers.BoolPtr(false)))
		assert.Equal(t, Null(), CopyArbitraryValue((*bool)(nil)))
	})
	t.Run("int8", func(t *testing.T) {
		var n int8 = 1
		assert.Equal(t, Int(1), CopyArbitraryValue(n))
		assert.Equal(t, Int(1), CopyArbitraryValue(&n))
		assert.Equal(t, Null(), CopyArbitraryValue((*int8)(nil)))
	})
	t.Run("uint8", func(t *testing.T) {
		var n uint8 = 1
		assert.Equal(t, Int(1), CopyArbitraryValue(n))
		assert.Equal(t, Int(1), CopyArbitraryValue(&n))
		assert.Equal(t, Null(), CopyArbitraryValue((*uint8)(nil)))
	})
	t.Run("int16", func(t *testing.T) {
		var n int16 = 1
		assert.Equal(t, Int(1), CopyArbitraryValue(n))
		assert.Equal(t, Int(1), CopyArbitraryValue(&n))
		assert.Equal(t, Null(), CopyArbitraryValue((*int16)(nil)))
	})
	t.Run("uint16", func(t *testing.T) {
		var n uint16 = 1
		assert.Equal(t, Int(1), CopyArbitraryValue(n))
		assert.Equal(t, Int(1), CopyArbitraryValue(&n))
		assert.Equal(t, Null(), CopyArbitraryValue((*uint16)(nil)))
	})
	t.Run("int", func(t *testing.T) {
		var n int = 1
		assert.Equal(t, Int(1), CopyArbitraryValue(n))
		assert.Equal(t, Int(1), CopyArbitraryValue(&n))
		assert.Equal(t, Null(), CopyArbitraryValue((*int)(nil)))
	})
	t.Run("uint", func(t *testing.T) {
		var n uint = 1
		assert.Equal(t, Int(1), CopyArbitraryValue(n))
		assert.Equal(t, Int(1), CopyArbitraryValue(&n))
		assert.Equal(t, Null(), CopyArbitraryValue((*uint)(nil)))
	})
	t.Run("int32", func(t *testing.T) {
		var n int32 = 1
		assert.Equal(t, Int(1), CopyArbitraryValue(n))
		assert.Equal(t, Int(1), CopyArbitraryValue(&n))
		assert.Equal(t, Null(), CopyArbitraryValue((*int32)(nil)))
	})
	t.Run("uint32", func(t *testing.T) {
		var n uint32 = 1
		assert.Equal(t, Int(1), CopyArbitraryValue(n))
		assert.Equal(t, Int(1), CopyArbitraryValue(&n))
		assert.Equal(t, Null(), CopyArbitraryValue((*uint32)(nil)))
	})
	t.Run("float32", func(t *testing.T) {
		var n float32 = 2.5
		assert.Equal(t, Float64(2.5), CopyArbitraryValue(n))
		assert.Equal(t, Float64(2.5), CopyArbitraryValue(&n))
		assert.Equal(t, Null(), CopyArbitraryValue((*float32)(nil)))
	})
	t.Run("float64", func(t *testing.T) {
		var n float64 = 2.5
		assert.Equal(t, Float64(2.5), CopyArbitraryValue(n))
		assert.Equal(t, Float64(2.5), CopyArbitraryValue(&n))
		assert.Equal(t, Null(), CopyArbitraryValue((*float64)(nil)))
	})
	t.Run("string", func(t *testing.T) {
		s := "x"
		assert.Equal(t, String(s), CopyArbitraryValue(s))
		assert.Equal(t, String(s), CopyArbitraryValue(&s))
		assert.Equal(t, Null(), CopyArbitraryValue((*string)(nil)))
	})
	t.Run("[]interface{}", func(t *testing.T) {
		a := []interface{}{2, []interface{}{"x"}}
		av := ArrayOf(Int(2), ArrayOf(String("x")))
		assert.Equal(t, av, CopyArbitraryValue(a))
		assert.Equal(t, av, CopyArbitraryValue(&a))
		assert.Equal(t, Null(), CopyArbitraryValue((*[]interface{})(nil)))
	})
	t.Run("[]Value", func(t *testing.T) {
		a := []Value{Int(2), ArrayOf(String("x"))}
		av := ArrayOf(Int(2), ArrayOf(String("x")))
		assert.Equal(t, av, CopyArbitraryValue(a))
		assert.Equal(t, av, CopyArbitraryValue(&a))
		assert.Equal(t, Null(), CopyArbitraryValue((*[]Value)(nil)))
	})
	t.Run("map[string]interface{}", func(t *testing.T) {
		m := map[string]interface{}{"x": []interface{}{2}}
		mv := ObjectBuild().Set("x", ArrayOf(Int(2))).Build()
		assert.Equal(t, mv, CopyArbitraryValue(m))
		assert.Equal(t, mv, CopyArbitraryValue(&m))
		assert.Equal(t, Null(), CopyArbitraryValue((*map[string]interface{})(nil)))
	})
	t.Run("map[string]Value", func(t *testing.T) {
		m := map[string]Value{"x": ArrayOf(Int(2))}
		mv := ObjectBuild().Set("x", ArrayOf(Int(2))).Build()
		assert.Equal(t, mv, CopyArbitraryValue(m))
		assert.Equal(t, mv, CopyArbitraryValue(&m))
		assert.Equal(t, Null(), CopyArbitraryValue((*map[string]Value)(nil)))
	})
	t.Run("arbitrary struct", func(t *testing.T) {
		s := struct {
			X int `json:"x"`
		}{X: 2}
		v := CopyArbitraryValue(s)
		assert.Equal(t, ObjectBuild().Set("x", Int(2)).Build(), v)
	})
	t.Run("raw", func(t *testing.T) {
		j := json.RawMessage("[3]")
		jv := Raw(json.RawMessage("[3]"))
		assert.Equal(t, jv, CopyArbitraryValue(j))
		assert.Equal(t, jv, CopyArbitraryValue(&j))
		assert.Equal(t, Null(), CopyArbitraryValue((*json.RawMessage)(nil)))
	})
}

func TestConvertPrimitivesToArbitraryValue(t *testing.T) {
	assert.Nil(t, Null().AsArbitraryValue())
	assert.Equal(t, true, Bool(true).AsArbitraryValue())
	assert.Equal(t, false, Bool(false).AsArbitraryValue())
	assert.Equal(t, float64(2), Int(2).AsArbitraryValue())
	assert.Equal(t, "x", String("x").AsArbitraryValue())
	assert.Equal(t, json.RawMessage("[3]"), Raw(json.RawMessage("[3]")).AsArbitraryValue())
}

func TestEqualPrimitives(t *testing.T) {
	valueFns := []func() Value{
		func() Value { return Null() },
		func() Value { return Bool(false) },
		func() Value { return Bool(true) },
		func() Value { return Int(1) },
		func() Value { return Float64(2.5) },
		func() Value { return String("") },
		func() Value { return String("1") },
		func() Value { return Raw(json.RawMessage("1")) },
	}
	for i, fn0 := range valueFns {
		v0 := fn0()
		for j, fn1 := range valueFns {
			v1 := fn1()
			if i == j {
				valuesShouldBeEqual(t, v0, v1)
			} else {
				valuesShouldNotBeEqual(t, v0, v1)
			}
		}
	}
}

func valuesShouldBeEqual(t *testing.T, value0 Value, value1 Value) {
	assert.True(t, value0.Equal(value1), "%s should equal %s", value0, value1)
	assert.True(t, value1.Equal(value0), "%s should equal %s conversely", value1, value0)
}

func valuesShouldNotBeEqual(t *testing.T, value0 Value, value1 Value) {
	assert.False(t, value0.Equal(value1), "%s should not equal %s", value0, value1)
	assert.False(t, value1.Equal(value0), "%s should not equal %s", value1, value0)
}

func TestValueWithInvalidType(t *testing.T) {
	// Application code has no way to construct a Value like this, but we'll still prove
	// that we would handle it gracefully if we did it somehow
	v := Value{valueType: ValueType(99)}

	assert.False(t, v.IsNull())
	assert.False(t, v.IsNumber())
	assert.False(t, v.IsInt())
	assert.False(t, v.BoolValue())
	assert.Equal(t, 0, v.IntValue())
	assert.Equal(t, float64(0), v.Float64Value())
	assert.Equal(t, "", v.StringValue())
	assert.Equal(t, OptionalString{}, v.AsOptionalString())
	assert.Equal(t, 0, v.Count())
	assert.Equal(t, Null(), v.GetByIndex(0))
	assert.Equal(t, Null(), v.GetByKey("x"))
	assert.Nil(t, v.AsArbitraryValue())
	assert.Nil(t, v.AsRaw())
}

func TestValueAsPointer(t *testing.T) {
	v := String("value")
	assert.Equal(t, &v, v.AsPointer())

	assert.Nil(t, Null().AsPointer())
}

func TestConvertArbitraryValueThatFailsToSerialize(t *testing.T) {
	v := CopyArbitraryValue(unserializableValue{})
	assert.Equal(t, Null(), v)
}

type unserializableValue struct{}

func (u unserializableValue) MarshalJSON() ([]byte, error) {
	return nil, errors.New("no")
}
