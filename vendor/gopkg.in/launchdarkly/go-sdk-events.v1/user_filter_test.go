package ldevents

import (
	"encoding/json"
	"sort"
	"testing"

	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream"

	"github.com/stretchr/testify/assert"

	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

var optionalStringSetters = map[lduser.UserAttribute]func(lduser.UserBuilder, string) lduser.UserBuilderCanMakeAttributePrivate{
	lduser.SecondaryKeyAttribute: lduser.UserBuilder.Secondary,
	lduser.IPAttribute:           lduser.UserBuilder.IP,
	lduser.CountryAttribute:      lduser.UserBuilder.Country,
	lduser.EmailAttribute:        lduser.UserBuilder.Email,
	lduser.FirstNameAttribute:    lduser.UserBuilder.FirstName,
	lduser.LastNameAttribute:     lduser.UserBuilder.LastName,
	lduser.AvatarAttribute:       lduser.UserBuilder.Avatar,
	lduser.NameAttribute:         lduser.UserBuilder.Name,
}

const customAttrName1 = "thing1"
const customAttrName2 = "thing2"

var customAttrValue1 = ldvalue.String("value1")
var customAttrValue2 = ldvalue.String("value2")

func buildUserWithAllAttributes() lduser.UserBuilder {
	return lduser.NewUserBuilder("user-key").
		FirstName("sam").
		LastName("smith").
		Name("sammy").
		Country("freedonia").
		Avatar("my-avatar").
		IP("123.456.789").
		Email("me@example.com").
		Secondary("abcdef").
		Anonymous(true).
		Custom(customAttrName1, customAttrValue1).
		Custom(customAttrName2, customAttrValue2)
}

func getAllPrivatableAttributeNames() []string {
	ret := []string{customAttrName1, customAttrName2}
	for a := range optionalStringSetters {
		ret = append(ret, string(a))
	}
	sort.Strings(ret)
	return ret
}

func makeUserOutput(uf userFilter, u EventUser) ldvalue.Value {
	var b jsonstream.JSONBuffer
	uf.writeUser(&b, u)
	bytes, err := b.Get()
	if err != nil {
		panic(err)
	}
	var v ldvalue.Value
	if err := json.Unmarshal(bytes, &v); err != nil {
		panic(err)
	}
	return v
}

func verifyUserHasPrivateTopLevelAttributeFilteredOut(
	t *testing.T,
	attr lduser.UserAttribute,
	filteredUserJSON ldvalue.Value,
	unfilteredUserJSON ldvalue.Value,
) {
	assert.Equal(t,
		ldvalue.ArrayBuild().Add(ldvalue.String(string(attr))).Build(),
		filteredUserJSON.GetByKey("privateAttrs"),
	)

	// One top-level attribute has been removed (the private one), but one has been added (privateAttrs)
	assert.Equal(t, unfilteredUserJSON.Count(), filteredUserJSON.Count())
	unfilteredUserJSON.Enumerate(func(i int, key string, value ldvalue.Value) bool {
		v1, found := filteredUserJSON.TryGetByKey(key)
		if key == string(attr) {
			assert.False(t, found, key)
		} else {
			assert.Equal(t, value, v1, key)
		}
		return true
	})
}

func verifyUserHasPrivateCustomAttributeFilteredOut(
	t *testing.T,
	attr lduser.UserAttribute,
	filteredUserJSON ldvalue.Value,
	unfilteredUserJSON ldvalue.Value,
) {
	assert.Equal(t,
		ldvalue.ArrayBuild().Add(ldvalue.String(string(attr))).Build(),
		filteredUserJSON.GetByKey("privateAttrs"),
	)

	// One top-level attribute has been added (privateAttrs)
	assert.Equal(t, unfilteredUserJSON.Count()+1, filteredUserJSON.Count())
	unfilteredUserJSON.Enumerate(func(i int, key string, value ldvalue.Value) bool {
		v1, _ := filteredUserJSON.TryGetByKey(key)
		if key == "custom" {
			assert.Equal(t, value.Count()-1, v1.Count())
			value.Enumerate(func(j int, key1 string, value1 ldvalue.Value) bool {
				v2, found := v1.TryGetByKey(key1)
				if key1 == string(attr) {
					assert.False(t, found, key1)
				} else {
					assert.Equal(t, value1, v2, key1)
				}
				return true
			})
		} else {
			assert.Equal(t, value, v1, key)
		}
		return true
	})
}

func TestWriteUserWithNoFiltering(t *testing.T) {
	expectUserSerializationUnchanged := func(t *testing.T, u lduser.User) {
		expectedJSON, _ := u.MarshalJSON()
		var expectedValue ldvalue.Value
		if err := json.Unmarshal(expectedJSON, &expectedValue); err != nil {
			panic(err)
		}
		filter := newUserFilter(epDefaultConfig)
		v := makeUserOutput(filter, EventUser{u, nil})
		assert.Equal(t, expectedValue, v)
	}

	t.Run("user with no attributes", func(t *testing.T) {
		expectUserSerializationUnchanged(t, lduser.NewUser("user-key"))
	})

	t.Run("user with all attributes", func(t *testing.T) {
		expectUserSerializationUnchanged(t, buildUserWithAllAttributes().Build())
	})
}

func TestWriteUserWithPerUserPrivateAttributes(t *testing.T) {
	filter := newUserFilter(epDefaultConfig)
	fu0 := makeUserOutput(filter, EventUser{buildUserWithAllAttributes().Build(), nil})

	for attr, setter := range optionalStringSetters {
		t.Run(string(attr), func(t *testing.T) {
			builder := buildUserWithAllAttributes()
			setter(builder, "private-value").AsPrivateAttribute()
			u1 := EventUser{builder.Build(), nil}
			fu1 := makeUserOutput(filter, u1)
			verifyUserHasPrivateTopLevelAttributeFilteredOut(t, attr, fu1, fu0)
		})
	}

	t.Run("custom", func(t *testing.T) {
		u1 := buildUserWithAllAttributes().
			Custom(customAttrName1, customAttrValue1).AsPrivateAttribute().
			Build()
		fu1 := makeUserOutput(filter, EventUser{u1, nil})

		verifyUserHasPrivateCustomAttributeFilteredOut(t, customAttrName1, fu1, fu0)
	})
}

func TestWriteUserWithGlobalPrivateAttributes(t *testing.T) {
	filter0 := newUserFilter(epDefaultConfig)
	u := EventUser{buildUserWithAllAttributes().Build(), nil}
	fu0 := makeUserOutput(filter0, u)

	for attr := range optionalStringSetters {
		t.Run(string(attr), func(t *testing.T) {
			config := epDefaultConfig
			config.PrivateAttributeNames = []lduser.UserAttribute{attr}
			filter1 := newUserFilter(config)
			fu1 := makeUserOutput(filter1, u)
			verifyUserHasPrivateTopLevelAttributeFilteredOut(t, attr, fu1, fu0)
		})
	}
	t.Run("custom", func(t *testing.T) {
		config := epDefaultConfig
		config.PrivateAttributeNames = []lduser.UserAttribute{lduser.UserAttribute(customAttrName1)}
		filter1 := newUserFilter(config)
		fu1 := makeUserOutput(filter1, u)
		verifyUserHasPrivateCustomAttributeFilteredOut(t, customAttrName1, fu1, fu0)
	})
	t.Run("allAttributesPrivate", func(t *testing.T) {
		config := epDefaultConfig
		config.AllAttributesPrivate = true
		filter1 := newUserFilter(config)
		fu1 := makeUserOutput(filter1, u)
		assert.Equal(t, 3, fu1.Count())
		assert.Equal(t, ldvalue.String(u.GetKey()), fu1.GetByKey("key"))
		assert.Equal(t, ldvalue.Bool(true), fu1.GetByKey("anonymous"))
		assert.Equal(t, len(getAllPrivatableAttributeNames()), fu1.GetByKey("privateAttrs").Count())
	})
}

func TestPrefilteredAttributesAreUsedUnchanged(t *testing.T) {
	config := epDefaultConfig
	config.AllAttributesPrivate = true // this should be ignored
	user := lduser.NewUserBuilder("user-key").Name("me").Build()
	eventUser := EventUser{User: user, AlreadyFilteredAttributes: []string{"firstName"}}
	filter := newUserFilter(config)
	fu := makeUserOutput(filter, eventUser)
	assert.Equal(t,
		ldvalue.ObjectBuild().
			Set("key", ldvalue.String(user.GetKey())).
			Set("name", user.GetName().AsValue()).
			Set("privateAttrs", ldvalue.ArrayBuild().Add(ldvalue.String("firstName")).Build()).
			Build(),
		fu)
}

func TestEmptyListOfPrefilteredAttributesIsUsedUnchanged(t *testing.T) {
	// This tests that setting AlreadyFilteredAttributes to an empty slice, unlike leaving it nil, is treated as
	// a sign that the user was already filtered and did not have any private attributes.
	config := epDefaultConfig
	config.AllAttributesPrivate = true // this should be ignored
	user := lduser.NewUserBuilder("user-key").Name("me").Build()
	eventUser := EventUser{User: user, AlreadyFilteredAttributes: []string{}}
	filter := newUserFilter(config)
	fu := makeUserOutput(filter, eventUser)
	assert.Equal(t,
		ldvalue.ObjectBuild().
			Set("key", ldvalue.String(user.GetKey())).
			Set("name", user.GetName().AsValue()).
			Build(),
		fu)
}
