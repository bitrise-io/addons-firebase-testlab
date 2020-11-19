package ldvalue

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayOf(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	value := ArrayOf(item0, item1)

	assert.Equal(t, ArrayType, value.Type())
	assert.Equal(t, 2, value.Count())

	assert.False(t, value.IsNull())
	assert.False(t, value.IsNumber())
	assert.False(t, value.IsInt())

	assert.False(t, value.BoolValue())
	assert.Equal(t, 0, value.IntValue())
	assert.Equal(t, float64(0), value.Float64Value())
	assert.Equal(t, "", value.StringValue())
	assert.Equal(t, OptionalString{}, value.AsOptionalString())
}

func TestArrayBuild(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	builder := ArrayBuild().Add(item0).Add(item1)
	value := builder.Build()

	assert.Equal(t, ArrayType, value.Type())
	assert.Equal(t, 2, value.Count())
	assert.Equal(t, ArrayOf(item0, item1), value)

	item2 := Bool(true)
	builder.Add(item2)
	valueAfterModifyingBuilder := builder.Build()

	assert.Equal(t, ArrayType, valueAfterModifyingBuilder.Type())
	assert.Equal(t, 3, valueAfterModifyingBuilder.Count())
	assert.Equal(t, item2, valueAfterModifyingBuilder.GetByIndex(2))

	assert.Equal(t, 2, value.Count()) // verifies builder's copy-on-write behavior

	assert.Equal(t, ArrayOf(), ArrayBuild().Build())
}

func TestArrayGetByIndex(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	value := ArrayOf(item0, item1)

	assert.Equal(t, item0, value.GetByIndex(0))
	assert.Equal(t, item1, value.GetByIndex(1))
	assert.Equal(t, Null(), value.GetByIndex(-1))
	assert.Equal(t, Null(), value.GetByIndex(2))

	item, ok := value.TryGetByIndex(0)
	assert.True(t, ok)
	assert.Equal(t, item0, item)
	item, ok = value.TryGetByIndex(2)
	assert.False(t, ok)
	assert.Equal(t, Null(), item)
}

func TestUsingArrayMethodsForNonArrayValue(t *testing.T) {
	values := []Value{
		Null(),
		Bool(true),
		Int(1),
		Float64(2.5),
		String(""),
		Raw(json.RawMessage("1")),
	}
	for _, v := range values {
		t.Run(v.String(), func(t *testing.T) {
			assert.Equal(t, 0, v.Count())
			assert.Equal(t, Null(), v.GetByIndex(0))
			_, ok := v.TryGetByIndex(0)
			assert.False(t, ok)
		})
	}
}

func TestObjectBuild(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	b := ObjectBuild().Set("item0", item0).Set("item1", item1)
	value := b.Build()

	assert.Equal(t, ObjectType, value.Type())
	assert.Equal(t, 2, value.Count())
	keys := value.Keys()
	sort.Strings(keys)
	assert.Equal(t, []string{"item0", "item1"}, keys)

	item0x := Bool(true)
	b.Set("item0", item0x)
	valueAfterModifyingBuilder := b.Build()
	assert.Equal(t, item0x, valueAfterModifyingBuilder.GetByKey("item0"))
	assert.Equal(t, item0, value.GetByKey("item0")) // verifies builder's copy-on-write behavior

	assert.False(t, value.IsNull())
	assert.False(t, value.IsNumber())
	assert.False(t, value.IsInt())

	assert.False(t, value.BoolValue())
	assert.Equal(t, 0, value.IntValue())
	assert.Equal(t, float64(0), value.Float64Value())
	assert.Equal(t, "", value.StringValue())
	assert.Equal(t, OptionalString{}, value.AsOptionalString())
}

func TestObjectGetByKey(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	value := ObjectBuild().Set("item0", item0).Set("item1", item1).Build()

	assert.Equal(t, item0, value.GetByKey("item0"))
	assert.Equal(t, item1, value.GetByKey("item1"))
	assert.Equal(t, Null(), value.GetByKey("bad-key"))

	item, ok := value.TryGetByKey("item0")
	assert.True(t, ok)
	assert.Equal(t, item0, item)
	item, ok = value.TryGetByKey("bad-key")
	assert.False(t, ok)
	assert.Equal(t, Null(), item)
}

func TestUsingObjectMethodsForNonObjectValue(t *testing.T) {
	values := []Value{
		Null(),
		Bool(true),
		Int(1),
		Float64(2.5),
		String(""),
		Raw(json.RawMessage("1")),
	}
	for _, v := range values {
		t.Run(v.String(), func(t *testing.T) {
			assert.Nil(t, v.Keys())
			assert.Equal(t, Null(), v.GetByKey("x"))
			_, ok := v.TryGetByKey("x")
			assert.False(t, ok)
		})
	}
}

func TestConvertComplexTypesFromArbitraryValue(t *testing.T) {
	t.Run("[]interface{}", func(t *testing.T) {
		v := CopyArbitraryValue([]interface{}{2, []interface{}{"x"}})
		assert.Equal(t, ArrayOf(Int(2), ArrayOf(String("x"))), v)
	})
	t.Run("[]Value", func(t *testing.T) {
		v := CopyArbitraryValue([]Value{Int(2), ArrayOf(String("x"))})
		assert.Equal(t, ArrayOf(Int(2), ArrayOf(String("x"))), v)
	})
	t.Run("map[string]interface{}", func(t *testing.T) {
		v := CopyArbitraryValue(map[string]interface{}{"x": []interface{}{2}})
		assert.Equal(t, ObjectBuild().Set("x", ArrayOf(Int(2))).Build(), v)
	})
	t.Run("map[string]Value", func(t *testing.T) {
		v := CopyArbitraryValue(map[string]Value{"x": ArrayOf(Int(2))})
		assert.Equal(t, ObjectBuild().Set("x", ArrayOf(Int(2))).Build(), v)
	})
	t.Run("arbitrary struct", func(t *testing.T) {
		s := struct {
			X int `json:"x"`
		}{2}
		v := CopyArbitraryValue(s)
		assert.Equal(t, ObjectBuild().Set("x", Int(2)).Build(), v)
	})
}

func TestConvertComplexTypesToArbitraryValue(t *testing.T) {
	t.Run("array", func(t *testing.T) {
		v := ArrayOf(Int(2), ArrayOf(String("x")))
		expected := []interface{}{float64(2), []interface{}{"x"}}
		assert.Equal(t, expected, v.AsArbitraryValue())
	})
	t.Run("object", func(t *testing.T) {
		v := ObjectBuild().Set("x", ArrayOf(Int(2))).Build()
		expected := map[string]interface{}{"x": []interface{}{float64(2)}}
		assert.Equal(t, expected, v.AsArbitraryValue())
	})
}

func TestConvertComplexTypesFromArbitraryValueAndBackAgain(t *testing.T) {
	t.Run("map[string]interface{}", func(t *testing.T) {
		mapValue0 := map[string]interface{}{"x": []interface{}{"b"}}
		v := CopyArbitraryValue(mapValue0)
		mapValue1 := v.AsArbitraryValue()
		assert.Equal(t, mapValue0, mapValue1)
		// Verify that the map was deep-copied
		mapValue0["x"].([]interface{})[0] = "c"
		assert.NotEqual(t, mapValue0, mapValue1)
	})
	t.Run("[]interface{}", func(t *testing.T) {
		sliceValue0 := []interface{}{[]interface{}{"b"}}
		v := CopyArbitraryValue(sliceValue0)
		sliceValue1 := v.AsArbitraryValue()
		assert.Equal(t, sliceValue0, sliceValue1)
		// Verify that the slice was deep-copied
		sliceValue0[0].([]interface{})[0] = "c"
		assert.NotEqual(t, sliceValue0, sliceValue1)
	})
}

func TestEqualComplexTypes(t *testing.T) {
	valueFns := []func() Value{
		func() Value { return Null() },
		func() Value { return Bool(false) },
		func() Value { return ArrayOf() },
		func() Value { return ArrayOf(Int(1)) },
		func() Value { return ArrayOf(Int(2)) },
		func() Value { return ArrayOf(Int(1), ArrayOf(String("a"))) },
		func() Value { return ArrayOf(Int(1), ArrayOf(String("a"), String("b"))) },
		func() Value { return ObjectBuild().Build() },
		func() Value { return ObjectBuild().Set("a", Int(1)).Build() },
		func() Value { return ObjectBuild().Set("a", Int(2)).Build() },
		func() Value { return ObjectBuild().Set("a", Int(1)).Set("b", Int(1)).Build() },
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

type enumerateParams struct {
	index int
	key   string
	value Value
}

func recordEnumerateCalls(value Value, stopFn func(enumerateParams) bool) []enumerateParams {
	ret := []enumerateParams{}
	value.Enumerate(func(index int, key string, v Value) bool {
		params := enumerateParams{index, key, v}
		ret = append(ret, params)
		if stopFn != nil && stopFn(params) {
			return false
		}
		return true
	})
	return ret
}

func TestEnumerateSimpleTypes(t *testing.T) {
	values := []Value{Bool(true), Int(1), String("x")}
	for _, value := range values {
		calls := recordEnumerateCalls(value, nil)
		assert.Equal(t, []enumerateParams{enumerateParams{0, "", value}}, calls)
	}
	assert.Equal(t, []enumerateParams{}, recordEnumerateCalls(Null(), nil))
}

func TestEnumerateArray(t *testing.T) {
	assert.Equal(t, []enumerateParams{}, recordEnumerateCalls(ArrayOf(), nil))

	assert.Equal(t, []enumerateParams{
		enumerateParams{0, "", Int(1)},
		enumerateParams{1, "", String("a")},
	}, recordEnumerateCalls(ArrayOf(Int(1), String("a")), nil))

	assert.Equal(t, []enumerateParams{
		enumerateParams{0, "", Int(1)},
	}, recordEnumerateCalls(ArrayOf(Int(1), String("a")), func(p enumerateParams) bool {
		return p.index == 0
	}))
}

func TestEnumerateObject(t *testing.T) {
	assert.Equal(t, []enumerateParams{}, recordEnumerateCalls(ObjectBuild().Build(), nil))

	o1 := ObjectBuild().Set("a", Int(1)).Set("b", Int(2)).Build()
	e1 := recordEnumerateCalls(o1, nil)
	sort.Slice(e1, func(i, j int) bool { return e1[i].key < e1[j].key })
	assert.Equal(t, []enumerateParams{
		enumerateParams{0, "a", Int(1)},
		enumerateParams{0, "b", Int(2)},
	}, e1)

	e2 := recordEnumerateCalls(o1, func(p enumerateParams) bool { return true })
	assert.Len(t, e2, 1)
}

func TestTransformSimpleTypes(t *testing.T) {
	fnDropAllValues := func(index int, key string, value Value) (Value, bool) {
		return String("should ignore this value since the second value is false"), false
	}
	fnWrapInArray := func(index int, key string, value Value) (Value, bool) {
		return ArrayOf(value), true
	}
	values := []Value{Bool(true), Int(1), String("x")}
	for _, value := range values {
		assert.Equal(t, Null(), value.Transform(fnDropAllValues))
		assert.Equal(t, ArrayOf(value), value.Transform(fnWrapInArray))
	}
	assert.Equal(t, Null(), Null().Transform(fnDropAllValues))
	assert.Equal(t, Null(), Null().Transform(fnWrapInArray))
}

func TestTransformArray(t *testing.T) {
	fnNoChanges := func(index int, key string, value Value) (Value, bool) {
		return value, true
	}
	fnAbsoluteValuesAndNoOddNumbers := func(index int, key string, value Value) (Value, bool) {
		if value.IntValue()%2 == 1 {
			return value, false // first return value should be ignored since second one is false
		}
		if value.IntValue() < 0 {
			return Int(-value.IntValue()), true
		}
		return value, true
	}
	fnTransformUsingIndex := func(index int, key string, value Value) (Value, bool) {
		return String(fmt.Sprintf("%d=%s", index, value.StringValue())), true
	}

	array1 := ArrayOf(Int(2), Int(4), Int(6))
	array1a := array1.Transform(fnNoChanges)
	array1b := array1.Transform(fnAbsoluteValuesAndNoOddNumbers)
	// Should have no changes...
	assert.Equal(t, array1, array1a)
	assert.Equal(t, array1, array1b)
	// ...and should be wrapping the *same* slice, not a copy
	array1.immutableArrayValue[0] = Int(0)
	assert.Equal(t, array1, array1a)
	assert.Equal(t, array1, array1b)

	array2 := ArrayOf(Int(2), Int(4), Int(1), Int(-6))
	array2a := array2.Transform(fnNoChanges)
	array2b := array2.Transform(fnAbsoluteValuesAndNoOddNumbers)
	// array2a should have no changes, and should be wrapping the same slice
	assert.Equal(t, array2, array2a)
	array2.immutableArrayValue[0] = Int(0)
	assert.Equal(t, array2, array2a)
	// array2b should have a transformed slice
	assert.Equal(t, ArrayOf(Int(2), Int(4), Int(6)), array2b)

	// Same as the array2 tests, except that the first change is a modification, not a deletion
	array3 := ArrayOf(Int(2), Int(4), Int(-6), Int(1))
	array3a := array3.Transform(fnNoChanges)
	array3b := array3.Transform(fnAbsoluteValuesAndNoOddNumbers)
	// array3a should have no changes, and should be wrapping the same slice
	assert.Equal(t, array3, array3a)
	array3.immutableArrayValue[0] = Int(0)
	assert.Equal(t, array3, array3a)
	// array3b should have a transformed slice
	assert.Equal(t, ArrayOf(Int(2), Int(4), Int(6)), array3b)

	// Edge case where the very first element is dropped
	array4 := ArrayOf(Int(1), Int(2), Int(4))
	array4b := array4.Transform(fnAbsoluteValuesAndNoOddNumbers)
	assert.Equal(t, ArrayOf(Int(2), Int(4)), array4b)

	// Edge case where the only element is dropped
	array5 := ArrayOf(Int(1))
	assert.Equal(t, ArrayOf(), array5.Transform(fnAbsoluteValuesAndNoOddNumbers))

	// Transformation function that uses the index parameter
	array6 := ArrayOf(String("a"), String("b"))
	assert.Equal(t, ArrayOf(String("0=a"), String("1=b")), array6.Transform(fnTransformUsingIndex))
}

func TestTransformObject(t *testing.T) {
	fnNoChanges := func(index int, key string, value Value) (Value, bool) {
		return value, true
	}
	fnAbsoluteValuesAndNoOddNumbers := func(index int, key string, value Value) (Value, bool) {
		if value.IntValue()%2 == 1 {
			return value, false // first return value should be ignored since second one is false
		}
		if value.IntValue() < 0 {
			return Int(-value.IntValue()), true
		}
		return value, true
	}
	fnTransformUsingKey := func(index int, key string, value Value) (Value, bool) {
		return String(fmt.Sprintf("%s=%s", key, value.JSONString())), true
	}

	o1 := ObjectBuild().Set("a", Int(2)).Set("b", Int(4)).Set("c", Int(6)).Build()
	o1a := o1.Transform(fnNoChanges)
	o1b := o1.Transform(fnAbsoluteValuesAndNoOddNumbers)
	// Should have no changes...
	assert.Equal(t, o1, o1a)
	assert.Equal(t, o1, o1b)
	// ...and should be wrapping the *same* map, not a copy
	o1.immutableObjectValue["a"] = Int(0)
	assert.Equal(t, o1, o1a)
	assert.Equal(t, o1, o1b)

	o2 := ObjectBuild().Set("a", Int(2)).Set("b", Int(4)).Set("c", Int(1)).Set("d", Int(-6)).Build()
	o2a := o2.Transform(fnNoChanges)
	o2b := o2.Transform(fnAbsoluteValuesAndNoOddNumbers)
	// o2a should have no changes, and should be wrapping the same map
	assert.Equal(t, o2, o2a)
	o2.immutableObjectValue["a"] = Int(0)
	assert.Equal(t, o2, o2a)
	// o2b should have a transformed map
	assert.Equal(t, ObjectBuild().Set("a", Int(2)).Set("b", Int(4)).Set("d", Int(6)).Build(), o2b)

	// Edge case where the only element is dropped
	o3 := ObjectBuild().Set("a", Int(1)).Build()
	assert.Equal(t, ObjectBuild().Build(), o3.Transform(fnAbsoluteValuesAndNoOddNumbers))

	// Transformation function that uses the key parameter
	o4 := ObjectBuild().Set("a", Int(2)).Set("b", Int(4)).Build()
	assert.Equal(t, ObjectBuild().Set("a", String("a=2")).Set("b", String("b=4")).Build(),
		o4.Transform(fnTransformUsingKey))

	// Case where we guarantee that the first element we iterated through is *not* modified - map
	// iteration order is nondeterministic, we just want to verify that we've hit all code paths
	n := 0
	fnTransformUsingKeyButNotFirst := func(index int, key string, value Value) (Value, bool) {
		n++
		if n > 1 {
			return fnTransformUsingKey(index, key, value)
		}
		return value, true
	}
	o5 := o4.Transform(fnTransformUsingKeyButNotFirst)
	assert.NotEqual(t, o4, o5)
	assert.Equal(t, o4.Count(), o5.Count())
	if o5.GetByKey("a").IsNumber() {
		assert.Equal(t, Int(2), o5.GetByKey("a"))
		assert.Equal(t, String("b=4"), o5.GetByKey("b"))
	} else {
		assert.Equal(t, String("a=2"), o5.GetByKey("a"))
		assert.Equal(t, Int(4), o5.GetByKey("b"))
	}
}
