package lduser

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

var allBuiltInAttributes = []UserAttribute{
	KeyAttribute,
	SecondaryKeyAttribute,
	IPAttribute,
	CountryAttribute,
	EmailAttribute,
	FirstNameAttribute,
	LastNameAttribute,
	AvatarAttribute,
	NameAttribute,
	AnonymousAttribute,
}

var optionalStringGetters = map[UserAttribute]func(User) ldvalue.OptionalString{
	SecondaryKeyAttribute: User.GetSecondaryKey,
	IPAttribute:           User.GetIP,
	CountryAttribute:      User.GetCountry,
	EmailAttribute:        User.GetEmail,
	FirstNameAttribute:    User.GetFirstName,
	LastNameAttribute:     User.GetLastName,
	AvatarAttribute:       User.GetAvatar,
	NameAttribute:         User.GetName,
}

var optionalStringSetters = map[UserAttribute]func(UserBuilder, string) UserBuilderCanMakeAttributePrivate{
	SecondaryKeyAttribute: UserBuilder.Secondary,
	IPAttribute:           UserBuilder.IP,
	CountryAttribute:      UserBuilder.Country,
	EmailAttribute:        UserBuilder.Email,
	FirstNameAttribute:    UserBuilder.FirstName,
	LastNameAttribute:     UserBuilder.LastName,
	AvatarAttribute:       UserBuilder.Avatar,
	NameAttribute:         UserBuilder.Name,
}

func assertStringAttrNotSet(t *testing.T, a UserAttribute, user User) {
	assert.Equal(t, ldvalue.OptionalString{}, optionalStringGetters[a](user), "should not have had a value for %s", a)
	assert.Equal(t, ldvalue.Null(), user.GetAttribute(a), "should not have had a value for %s", a)
}

func assertStringPropertiesNotSet(t *testing.T, user User) {
	for a, _ := range optionalStringGetters {
		assertStringAttrNotSet(t, a, user)
	}
}

func getCustomAttrs(user User) []string {
	var ret []string
	user.GetAllCustom().Enumerate(func(i int, a string, v ldvalue.Value) bool {
		ret = append(ret, a)
		return true
	})
	sort.Strings(ret)
	return ret
}

func getPrivateAttrs(user User) []string {
	var ret []string
	for _, a := range allBuiltInAttributes {
		if user.IsPrivateAttribute(a) {
			ret = append(ret, string(a))
		}
	}
	user.GetAllCustom().Enumerate(func(i int, a string, v ldvalue.Value) bool {
		if user.IsPrivateAttribute(UserAttribute(a)) {
			ret = append(ret, a)
		}
		return true
	})
	sort.Strings(ret)
	return ret
}

func TestNewUser(t *testing.T) {
	user := NewUser("some-key")

	assert.Equal(t, "some-key", user.GetKey())
	assert.Equal(t, ldvalue.String("some-key"), user.GetAttribute(KeyAttribute))

	assertStringPropertiesNotSet(t, user)

	anon, found := user.GetAnonymousOptional()
	assert.False(t, anon)
	assert.False(t, found)
	assert.Equal(t, ldvalue.Null(), user.GetAttribute(AnonymousAttribute))

	assert.Nil(t, getCustomAttrs(user))
	assert.Nil(t, getPrivateAttrs(user))
	assert.False(t, user.HasPrivateAttributes())
}

func TestNewAnonymousUser(t *testing.T) {
	user := NewAnonymousUser("some-key")

	assert.Equal(t, "some-key", user.GetKey())
	assert.Equal(t, ldvalue.String("some-key"), user.GetAttribute(KeyAttribute))

	assertStringPropertiesNotSet(t, user)

	anon, found := user.GetAnonymousOptional()
	assert.True(t, anon)
	assert.True(t, found)
	assert.Equal(t, ldvalue.Bool(true), user.GetAttribute(AnonymousAttribute))

	assert.Nil(t, getCustomAttrs(user))
	assert.Nil(t, getPrivateAttrs(user))
	assert.False(t, user.HasPrivateAttributes())
}

func TestUserBuilderSetsOnlyKeyByDefault(t *testing.T) {
	user := NewUserBuilder("some-key").Build()

	assert.Equal(t, "some-key", user.GetKey())
	assert.Equal(t, ldvalue.String("some-key"), user.GetAttribute(KeyAttribute))

	assertStringPropertiesNotSet(t, user)

	anon, found := user.GetAnonymousOptional()
	assert.False(t, anon)
	assert.False(t, found)
	assert.Equal(t, ldvalue.Null(), user.GetAttribute(AnonymousAttribute))

	assert.Nil(t, getCustomAttrs(user))
	assert.Nil(t, getPrivateAttrs(user))
	assert.False(t, user.HasPrivateAttributes())
}

func TestUserBuilderCanSetStringAttributes(t *testing.T) {
	for a, setter := range optionalStringSetters {
		t.Run(string(a), func(t *testing.T) {
			builder := NewUserBuilder("some-key")
			setter(builder, "value")
			user := builder.Build()

			assert.Equal(t, "some-key", user.GetKey())
			assert.Equal(t, ldvalue.String("some-key"), user.GetAttribute(KeyAttribute))

			assert.Equal(t, ldvalue.NewOptionalString("value"), optionalStringGetters[a](user), a)
			assert.Equal(t, ldvalue.String("value"), user.GetAttribute(a))

			for a1, _ := range optionalStringSetters {
				if a1 != a {
					assertStringAttrNotSet(t, a1, user)
				}
			}
		})
	}
}

func TestUserBuilderCanSetAnonymous(t *testing.T) {
	user0 := NewUserBuilder("some-key").Build()
	assert.False(t, user0.GetAnonymous())
	value, ok := user0.GetAnonymousOptional()
	assert.False(t, ok)
	assert.False(t, value)
	assert.Equal(t, ldvalue.Null(), user0.GetAttribute(AnonymousAttribute))

	user1 := NewUserBuilder("some-key").Anonymous(true).Build()
	assert.True(t, user1.GetAnonymous())
	value, ok = user1.GetAnonymousOptional()
	assert.True(t, ok)
	assert.True(t, value)
	assert.Equal(t, ldvalue.Bool(true), user1.GetAttribute(AnonymousAttribute))

	user2 := NewUserBuilder("some-key").Anonymous(false).Build()
	assert.False(t, user2.GetAnonymous())
	value, ok = user2.GetAnonymousOptional()
	assert.True(t, ok)
	assert.False(t, value)
	assert.Equal(t, ldvalue.Bool(false), user2.GetAttribute(AnonymousAttribute))
}

func TestUserBuilderCanSetPrivateStringAttributes(t *testing.T) {
	for a, setter := range optionalStringSetters {
		t.Run(string(a), func(t *testing.T) {
			builder := NewUserBuilder("some-key")
			setter(builder, "value").AsPrivateAttribute()
			user := builder.Build()

			assert.Equal(t, "some-key", user.GetKey())

			assert.Equal(t, ldvalue.NewOptionalString("value"), optionalStringGetters[a](user))
			assert.Equal(t, ldvalue.String("value"), user.GetAttribute(a))

			for a1, _ := range optionalStringSetters {
				if a1 != a {
					assertStringAttrNotSet(t, a1, user)
				}
			}

			assert.Nil(t, getCustomAttrs(user))
			assert.Equal(t, []string{string(a)}, getPrivateAttrs(user))
			assert.True(t, user.HasPrivateAttributes())
		})
	}
}

func TestUserBuilderCanMakeAttributeNonPrivate(t *testing.T) {
	builder := NewUserBuilder("some-key")
	builder.Country("us").AsNonPrivateAttribute()
	builder.Email("e").AsPrivateAttribute()
	builder.Name("n").AsPrivateAttribute()
	builder.Email("f").AsNonPrivateAttribute()
	user := builder.Build()
	assert.Equal(t, "f", user.GetEmail().StringValue())
	assert.Equal(t, []string{"name"}, getPrivateAttrs(user))
	assert.True(t, user.HasPrivateAttributes())
}

func TestUserBuilderCanSetCustomAttributes(t *testing.T) {
	user := NewUserBuilder("some-key").Custom("first", ldvalue.Int(1)).Custom("second", ldvalue.String("two")).Build()

	value, ok := user.GetCustom("first")
	assert.True(t, ok)
	assert.Equal(t, 1, value.IntValue())

	value, ok = user.GetCustom("second")
	assert.True(t, ok)
	assert.Equal(t, "two", value.StringValue())

	value, ok = user.GetCustom("no")
	assert.False(t, ok)
	assert.Equal(t, ldvalue.Null(), value)

	assert.Equal(t, []string{"first", "second"}, getCustomAttrs(user))

	assert.Nil(t, getPrivateAttrs(user))
}

func TestUserBuilderCanSetAttributesAfterSettingAttributeThatCanBePrivate(t *testing.T) {
	// This tests that chaining methods off of UserBuilderCanMakeAttributePrivate works correctly.
	builder := NewUserBuilder("some-key").Name("original-name")
	builder.Key("new-key")
	builder.Anonymous(true)
	builder.Custom("thing", ldvalue.String("custom-value"))
	user := builder.Build()

	assert.Equal(t, "new-key", user.GetKey())
	assert.Equal(t, ldvalue.Bool(true), user.GetAttribute(AnonymousAttribute))
	assert.Equal(t, ldvalue.String("custom-value"), user.GetAttribute(UserAttribute("thing")))

	for a, setter := range optionalStringSetters {
		t.Run(string(a), func(t *testing.T) {
			builder := NewUserBuilder("some-key").Name("original-name")
			// builder is now a UserBuilderCanMakeAttributePrivate
			setter(builder, "value")
			user := builder.Build()

			assert.Equal(t, ldvalue.NewOptionalString("value"), optionalStringGetters[a](user))
			assert.Equal(t, ldvalue.String("value"), user.GetAttribute(a))
		})
	}
}

func TestUserBuilderGenericSetAttribute(t *testing.T) {
	t.Run("key", func(t *testing.T) {
		builder := NewUserBuilder("some-key")
		value := "value"

		builder.SetAttribute(KeyAttribute, ldvalue.String(value))
		assert.Equal(t, value, builder.Build().GetKey())

		builder.SetAttribute(KeyAttribute, ldvalue.Null())
		assert.Equal(t, value, builder.Build().GetKey())

		builder.SetAttribute(KeyAttribute, ldvalue.Bool(true))
		assert.Equal(t, value, builder.Build().GetKey())

		builder.SetAttribute(KeyAttribute, ldvalue.Int(1))
		assert.Equal(t, value, builder.Build().GetKey())

		builder.SetAttribute(KeyAttribute, ldvalue.String(value)).AsPrivateAttribute()
		assert.Len(t, getPrivateAttrs(builder.Build()), 0)
	})

	for a, getter := range optionalStringGetters {
		t.Run(string(a), func(t *testing.T) {
			builder := NewUserBuilder("some-key")
			valueStr := ldvalue.NewOptionalString("value")
			value := valueStr.AsValue()

			builder.SetAttribute(a, value)
			assert.Equal(t, valueStr, getter(builder.Build()))

			for a1, _ := range optionalStringGetters {
				if a1 != a {
					assertStringAttrNotSet(t, a1, builder.Build())
				}
			}

			assert.Len(t, getPrivateAttrs(builder.Build()), 0)

			builder.SetAttribute(a, ldvalue.Bool(true))
			assert.Equal(t, valueStr, getter(builder.Build()))

			builder.SetAttribute(a, ldvalue.Int(1))
			assert.Equal(t, valueStr, getter(builder.Build()))

			builder.SetAttribute(a, ldvalue.Null())
			assert.Equal(t, ldvalue.OptionalString{}, getter(builder.Build()))

			builder.SetAttribute(a, value).AsPrivateAttribute()

			assert.Equal(t, []string{string(a)}, getPrivateAttrs(builder.Build()))
		})
	}

	t.Run("anonymous", func(t *testing.T) {
		builder := NewUserBuilder("some-key")

		builder.SetAttribute(AnonymousAttribute, ldvalue.Bool(false))
		value, ok := builder.Build().GetAnonymousOptional()
		assert.True(t, ok)
		assert.False(t, value)
		assert.False(t, builder.Build().GetAnonymous())
		builder.SetAttribute(AnonymousAttribute, ldvalue.Int(1))
		assert.False(t, builder.Build().GetAnonymous())

		builder.SetAttribute(AnonymousAttribute, ldvalue.Bool(true))
		value, ok = builder.Build().GetAnonymousOptional()
		assert.True(t, ok)
		assert.True(t, value)
		assert.True(t, builder.Build().GetAnonymous())
		builder.SetAttribute(AnonymousAttribute, ldvalue.Int(1))
		assert.True(t, builder.Build().GetAnonymous())

		builder.SetAttribute(AnonymousAttribute, ldvalue.Null())
		assert.False(t, builder.Build().GetAnonymous())
		value, ok = builder.Build().GetAnonymousOptional()
		assert.False(t, ok)
		assert.False(t, value)

		builder.SetAttribute(AnonymousAttribute, ldvalue.Bool(true)).AsPrivateAttribute()
		assert.Len(t, getPrivateAttrs(builder.Build()), 0)
	})

	t.Run("custom", func(t *testing.T) {
		builder := NewUserBuilder("some-key")
		a := UserAttribute("thing")
		value := ldvalue.Int(2)

		builder.SetAttribute(a, value)
		assert.Equal(t, value, builder.Build().GetAttribute(a))

		assert.Len(t, getPrivateAttrs(builder.Build()), 0)

		builder.SetAttribute(a, ldvalue.Null())
		assert.Equal(t, ldvalue.Null(), builder.Build().GetAttribute(a))

		builder.SetAttribute(a, value).AsPrivateAttribute()
		assert.Equal(t, []string{string(a)}, getPrivateAttrs(builder.Build()))
	})
}

func TestEnumerateCustomAttributes(t *testing.T) {
	user := NewUserBuilder("some-key").Custom("first", ldvalue.Int(1)).Custom("second", ldvalue.String("two")).Build()
	m := make(map[string]ldvalue.Value)
	user.GetAllCustom().Enumerate(func(i int, a string, v ldvalue.Value) bool {
		m[a] = v
		return true
	})
	assert.Equal(t, map[string]ldvalue.Value{"first": ldvalue.Int(1), "second": ldvalue.String("two")}, m)
}

func TestUserWithNoCustomAttributes(t *testing.T) {
	user := NewUser("some-key")

	assert.Nil(t, getCustomAttrs(user))

	value, ok := user.GetCustom("attr")
	assert.False(t, ok)
	assert.Equal(t, ldvalue.Null(), value)

	assert.Nil(t, getCustomAttrs(user))
}

func TestUserBuilderCanSetPrivateCustomAttributes(t *testing.T) {
	user := NewUserBuilder("some-key").Custom("first", ldvalue.Int(1)).AsPrivateAttribute().
		Custom("second", ldvalue.String("two")).Build()

	value, ok := user.GetCustom("first")
	assert.True(t, ok)
	assert.Equal(t, 1, value.IntValue())

	value, ok = user.GetCustom("second")
	assert.True(t, ok)
	assert.Equal(t, "two", value.StringValue())

	value, ok = user.GetCustom("no")
	assert.False(t, ok)
	assert.Equal(t, ldvalue.Null(), value)

	assert.Equal(t, []string{"first", "second"}, getCustomAttrs(user))

	assert.Equal(t, []string{"first"}, getPrivateAttrs(user))
	assert.True(t, user.HasPrivateAttributes())
}

func TestUserBuilderCanCopyFromExistingUserWithOnlyKey(t *testing.T) {
	user0 := NewUser("some-key")
	user1 := NewUserBuilderFromUser(user0).Build()

	assert.Equal(t, "some-key", user1.GetKey())

	assertStringPropertiesNotSet(t, user1)
	assert.Nil(t, getCustomAttrs(user1))
	assert.Nil(t, getPrivateAttrs(user1))
	assert.False(t, user1.HasPrivateAttributes())
}

func TestUserBuilderCanCopyFromExistingUserWithAllAttributes(t *testing.T) {
	user0 := newUserBuilderWithAllPropertiesSet("some-key").Build()
	user1 := NewUserBuilderFromUser(user0).Build()
	assert.Equal(t, user0, user1)
}

func TestGetUnknownAttribute(t *testing.T) {
	assert.Equal(t, ldvalue.Null(), NewUser("some-key").GetAttribute(UserAttribute("no-such-thing")))
}

func TestUserEqualsComparesAllAttributes(t *testing.T) {
	shouldEqual := func(a User, b User) {
		assert.True(t, b.Equal(a), "%s should equal %s", b, a)
		assert.True(t, a.Equal(b), "%s should equal %s", a, b)
	}
	shouldNotEqual := func(a User, b User) {
		assert.False(t, b.Equal(a), "%s should not equal %s", b, a)
		assert.False(t, a.Equal(b), "%s should not equal %s", a, b)
	}

	user0 := NewUser("some-key")
	assert.True(t, user0.Equal(user0), "%s should equal itself", user0)

	user1 := newUserBuilderWithAllPropertiesSet("some-key").Build()
	assert.True(t, user1.Equal(user1), "%s should equal itself", user1)
	user2 := NewUserBuilderFromUser(user1).Build()
	shouldEqual(user1, user2)

	for a, setter := range optionalStringSetters {
		builder3 := NewUserBuilderFromUser(user1)
		setter(builder3, "different-value")
		user3 := builder3.Build()
		shouldNotEqual(user1, user3)

		builder4 := NewUserBuilderFromUser(user1)
		setter(builder4, fmt.Sprintf("value-%s", a)).AsPrivateAttribute()
		user4 := builder4.Build()
		shouldNotEqual(user1, user4)
	}

	shouldNotEqual(user1, NewUserBuilderFromUser(user1).Key("other-key").Build())

	shouldNotEqual(user0, NewUserBuilderFromUser(user0).Anonymous(true).Build())
	shouldNotEqual(NewUserBuilderFromUser(user0).Anonymous(true).Build(), NewUserBuilderFromUser(user0).Anonymous(false).Build())

	// modifying an existing custom attribute
	shouldNotEqual(user1, NewUserBuilderFromUser(user1).Custom("thing1", ldvalue.String("other-value")).Build())

	// adding a new custom attribute
	shouldNotEqual(user1, NewUserBuilderFromUser(user1).Custom("thing3", ldvalue.String("other-value")).Build())

	// adding an extra private attribute
	shouldNotEqual(user1, NewUserBuilderFromUser(user1).Custom("thing1", ldvalue.String("value1")).AsPrivateAttribute().Build())

	// having the same number of private attributes, but not the same ones
	shouldNotEqual(
		NewUserBuilderFromUser(user1).
			Custom("thing3", ldvalue.Bool(true)).AsPrivateAttribute().
			Custom("thing4", ldvalue.Bool(true)).
			Build(),
		NewUserBuilderFromUser(user1).
			Custom("thing3", ldvalue.Bool(true)).
			Custom("thing4", ldvalue.Bool(true)).AsPrivateAttribute().
			Build())
}

func newUserBuilderWithAllPropertiesSet(key string) UserBuilder {
	builder := NewUserBuilder(key)
	for a, setter := range optionalStringSetters {
		setter(builder, fmt.Sprintf("value-%s", a))
	}
	builder.Anonymous(true)
	builder.Custom("thing1", ldvalue.String("value1"))
	builder.Custom("thing2", ldvalue.String("value2")).AsPrivateAttribute()
	return builder
}
