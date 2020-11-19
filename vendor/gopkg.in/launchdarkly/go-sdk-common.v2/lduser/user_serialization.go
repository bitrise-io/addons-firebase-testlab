package lduser

import (
	"encoding/json"
	"reflect"

	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

type missingKeyError struct{}

func (e missingKeyError) Error() string {
	return "User must have a key property"
}

// ErrMissingKey returns the standard error value that is used if you try to unmarshal a user from JSON
// and the "key" property is either absent or null. This is distinguished from other kinds of unmarshaling
// errors (such as trying to set a string property to a non-string value) in order to support use cases
// where incomplete data needs to be treated differently from malformed data.
//
// LaunchDarkly does allow a user to have an empty string ("") as a key in some cases, but this is
// discouraged since analytics events will not work properly without unique user keys.
func ErrMissingKey() error {
	return missingKeyError{}
}

// This temporary struct allows us to do JSON unmarshalling as efficiently as possible while not requiring
// the User's internal representation to be constrained by the behavior of json.Unmarshal.
type userForDeserialization struct {
	Key                   ldvalue.OptionalString `json:"key"`
	Secondary             ldvalue.OptionalString `json:"secondary"`
	IP                    ldvalue.OptionalString `json:"ip"`
	Country               ldvalue.OptionalString `json:"country"`
	Email                 ldvalue.OptionalString `json:"email"`
	FirstName             ldvalue.OptionalString `json:"firstName"`
	LastName              ldvalue.OptionalString `json:"lastName"`
	Avatar                ldvalue.OptionalString `json:"avatar"`
	Name                  ldvalue.OptionalString `json:"name"`
	Anonymous             ldvalue.OptionalBool   `json:"anonymous"`
	Custom                ldvalue.Value          `json:"custom"`
	PrivateAttributeNames []UserAttribute        `json:"privateAttributeNames"`
}

// String returns a simple string representation of a user.
//
// This currently uses the same JSON string representation as User.MarshalJSON(). Do not rely on this
// specific behavior of String(); it is intended for convenience in debugging.
func (u User) String() string {
	bytes, _ := json.Marshal(u)
	return string(bytes)
}

// MarshalJSON provides JSON serialization for User when using json.MarshalJSON.
//
// This is LaunchDarkly's standard JSON representation for user properties, in which all of the built-in
// user attributes are at the top level along with a "custom" property that is an object containing all of
// the custom attributes.
//
// In order for the representation to be as compact as possible, any top-level attributes for which no
// value has been set (as opposed to being set to an empty string) will be completely omitted, rather
// than including "attributeName":null in the JSON output. Similarly, if there are no custom attributes,
// there will be no "custom" property (rather than "custom":{}). This distinction does not matter to
// LaunchDarkly services-- they will treat an explicit null value in JSON data the same as an unset
// attribute, and treat an omitted "custom" the same as an empty "custom" map.
func (u User) MarshalJSON() ([]byte, error) {
	var buf jsonstream.JSONBuffer
	u.WriteToJSONBuffer(&buf)
	return buf.Get()
}

// UnmarshalJSON provides JSON deserialization for User when using json.UnmarshalJSON.
//
// This is LaunchDarkly's standard JSON representation for user properties, in which all of the built-in
// properties are at the top level along with a "custom" property that is an object containing all of
// the custom properties.
//
// Any property that is either completely omitted or has a null value is ignored and left in an unset
// state, except for "key". All users must have a key (even if it is ""), so an omitted or null "key"
// property causes the error ErrMissingKey().
//
// Trying to unmarshal any non-struct value, including a JSON null, into a User will return a
// json.UnmarshalTypeError. If you want to unmarshal optional user data that might be null, use *User
// instead of User.
func (u *User) UnmarshalJSON(data []byte) error {
	// Special handling here for a null value - json.Unmarshal will normally treat a null exactly like
	// "{}" when unmarshaling a struct. We don't want that, because it will produce a misleading
	// "missing key" error further down. Instead, just treat it as an invalid type.
	if string(data) == "null" {
		return &json.UnmarshalTypeError{Value: string(data), Type: reflect.TypeOf(u)}
	}
	var ufs userForDeserialization
	if err := json.Unmarshal(data, &ufs); err != nil {
		return err
	}
	if !ufs.Key.IsDefined() {
		return ErrMissingKey()
	}
	*u = User{
		key:       ufs.Key.StringValue(),
		secondary: ufs.Secondary,
		ip:        ufs.IP,
		country:   ufs.Country,
		email:     ufs.Email,
		firstName: ufs.FirstName,
		lastName:  ufs.LastName,
		avatar:    ufs.Avatar,
		name:      ufs.Name,
		anonymous: ufs.Anonymous,
		custom:    ufs.Custom,
	}
	if len(ufs.PrivateAttributeNames) > 0 {
		u.privateAttributes = make(map[UserAttribute]struct{})
		for _, a := range ufs.PrivateAttributeNames {
			u.privateAttributes[a] = struct{}{}
		}
	}
	return nil
}

// WriteToJSONBuffer provides JSON serialization for User with the jsonstream API.
//
// The JSON output format is identical to what is produced by json.Marshal, but this implementation is
// more efficient when building output with JSONBuffer. See the jsonstream package for more details.
func (u User) WriteToJSONBuffer(j *jsonstream.JSONBuffer) {
	j.BeginObject()
	j.WriteName("key")
	j.WriteString(u.key)
	maybeWriteStringProperty(j, "secondary", u.secondary)
	maybeWriteStringProperty(j, "ip", u.ip)
	maybeWriteStringProperty(j, "country", u.country)
	maybeWriteStringProperty(j, "email", u.email)
	maybeWriteStringProperty(j, "firstName", u.firstName)
	maybeWriteStringProperty(j, "lastName", u.lastName)
	maybeWriteStringProperty(j, "avatar", u.avatar)
	maybeWriteStringProperty(j, "name", u.name)
	if u.anonymous.IsDefined() {
		j.WriteName("anonymous")
		j.WriteBool(u.anonymous.BoolValue())
	}
	if u.custom.Count() > 0 {
		j.WriteName("custom")
		j.BeginObject()
		u.custom.Enumerate(func(i int, key string, value ldvalue.Value) bool {
			j.WriteName(key)
			value.WriteToJSONBuffer(j)
			return true
		})
		j.EndObject()
	}
	if len(u.privateAttributes) > 0 {
		j.WriteName("privateAttributeNames")
		j.BeginArray()
		for name := range u.privateAttributes {
			j.WriteString(string(name))
		}
		j.EndArray()
	}
	j.EndObject()
}

func maybeWriteStringProperty(j *jsonstream.JSONBuffer, name string, value ldvalue.OptionalString) {
	if value.IsDefined() {
		j.WriteName(name)
		j.WriteString(value.StringValue())
	}
}
