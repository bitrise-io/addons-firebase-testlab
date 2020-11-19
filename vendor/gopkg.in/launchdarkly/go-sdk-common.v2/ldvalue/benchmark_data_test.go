package ldvalue

var (
	benchmarkBoolValue          = true
	benchmarkBoolPointer        = &benchmarkBoolValue
	benchmarkOptBoolWithValue   = NewOptionalBool(benchmarkBoolValue)
	benchmarkIntValue           = 3333
	benchmarkIntPointer         = &benchmarkIntValue
	benchmarkOptIntWithValue    = NewOptionalInt(benchmarkIntValue)
	benchmarkStringValue        = "value"
	benchmarkStringPointer      = &benchmarkStringValue
	benchmarkOptStringWithValue = NewOptionalString(benchmarkStringValue)

	benchmarkSerializeNullValue   = Null()
	benchmarkSerializeBoolValue   = Bool(true)
	benchmarkSerializeIntValue    = Int(1000)
	benchmarkSerializeFloatValue  = Float64(1000.5)
	benchmarkSerializeStringValue = String("value")
	benchmarkSerializeArrayValue  = ArrayOf(String("a"), String("b"), String("c"))
	benchmarkSerializeObjectValue = ObjectBuild().Set("a", Int(1)).Set("b", Int(2)).Set("c", Int(3)).Build()

	benchmarkOptBoolResult   OptionalBool
	benchmarkOptIntResult    OptionalInt
	benchmarkOptStringResult OptionalString
	benchmarkStringResult    string
	benchmarkValueResult     Value
	benchmarkBoolResult      bool
	benchmarkIntResult       int
	benchmarkFloat64Result   float64
	benchmarkJSONResult      []byte
)
