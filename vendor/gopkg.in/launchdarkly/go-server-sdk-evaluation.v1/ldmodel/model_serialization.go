package ldmodel

// DataModelSerialization defines an encoding for SDK data model objects.
//
// For the default JSON encoding used by LaunchDarkly SDKs, use NewJSONDataModelSerialization.
type DataModelSerialization interface {
	// MarshalFeatureFlag converts a FeatureFlag into its serialized encoding.
	MarshalFeatureFlag(item FeatureFlag) ([]byte, error)
	// MarshalSegment converts a Segment into its serialized encoding.
	MarshalSegment(item Segment) ([]byte, error)
	// UnmarshalFeatureFlag attempts to convert a FeatureFlag from its serialized encoding.
	UnmarshalFeatureFlag(data []byte) (FeatureFlag, error)
	// UnmarshalFeatureFlag attempts to convert a FeatureFlag from its serialized encoding.
	UnmarshalSegment(data []byte) (Segment, error)
}

type jsonDataModelSerialization struct{}

// NewJSONDataModelSerialization provides the default JSON encoding for SDK data model objects.
//
// Always use this rather than relying on json.Marshal() and json.Unmarshal(). The data model
// structs are guaranteed to serialize and deserialize correctly with json.Marshal() and
// json.Unmarshal(), but JSONDataModelSerialization may be enhanced in the future to use a
// more efficient mechanism.
func NewJSONDataModelSerialization() DataModelSerialization {
	return jsonDataModelSerialization{}
}

func (s jsonDataModelSerialization) MarshalFeatureFlag(item FeatureFlag) ([]byte, error) {
	return marshalFeatureFlag(item)
}

func (s jsonDataModelSerialization) MarshalSegment(item Segment) ([]byte, error) {
	return marshalSegment(item)
}

func (s jsonDataModelSerialization) UnmarshalFeatureFlag(data []byte) (FeatureFlag, error) {
	return unmarshalFeatureFlag(data)
}

func (s jsonDataModelSerialization) UnmarshalSegment(data []byte) (Segment, error) {
	return unmarshalSegment(data)
}

// MarshalJSON overrides the default json.Marshal behavior to provide the same marshalling behavior that is
// used by NewJSONDataModelSerialization().
func (f FeatureFlag) MarshalJSON() ([]byte, error) {
	return marshalFeatureFlag(f)
}

// MarshalJSON overrides the default json.Marshal behavior to provide the same marshalling behavior that is
// used by NewJSONDataModelSerialization().
func (s Segment) MarshalJSON() ([]byte, error) {
	return marshalSegment(s)
}

// UnmarshalJSON overrides the default json.Unmarshal behavior to provide the same unmarshalling behavior that
// is used by NewJSONDataModelSerialization().
func (f *FeatureFlag) UnmarshalJSON(data []byte) error {
	result, err := unmarshalFeatureFlag(data)
	if err == nil {
		*f = result
	}
	return err
}

// UnmarshalJSON overrides the default json.Unmarshal behavior to provide the same unmarshalling behavior that
// is used by NewJSONDataModelSerialization().
func (s *Segment) UnmarshalJSON(data []byte) error {
	result, err := unmarshalSegment(data)
	if err == nil {
		*s = result
	}
	return err
}
