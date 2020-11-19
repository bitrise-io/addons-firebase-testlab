package ldevents

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldlog"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

type userFilter struct {
	allAttributesPrivate    bool
	globalPrivateAttributes []lduser.UserAttribute
	loggers                 ldlog.Loggers
	logUserKeyInErrors      bool
}

func newUserFilter(config EventsConfiguration) userFilter {
	return userFilter{
		allAttributesPrivate:    config.AllAttributesPrivate,
		globalPrivateAttributes: config.PrivateAttributeNames,
		loggers:                 config.Loggers,
		logUserKeyInErrors:      config.LogUserKeyInErrors,
	}
}

// Produces a JSON serialization of the user for event data and writes it to the specified JSONBuffer.
//
// If neither the configuration nor the user specifies any private attributes, then this is the same
// as the original user. Otherwise, it is a copy which may have some attributes removed (with the
// PrivateAttributes property set to a list of their names).
func (uf *userFilter) writeUser(b *jsonstream.JSONBuffer, user EventUser) {
	b.BeginObject()

	b.WriteName("key")
	b.WriteString(user.GetKey())

	var privateAttrsPreallocateArray [30]string // lets us avoid heap allocation in typical cases
	var privateAttrs = privateAttrsPreallocateArray[0:0]

	uf.maybeStringAttribute(b, &user, lduser.SecondaryKeyAttribute, user.GetSecondaryKey(), &privateAttrs)
	uf.maybeStringAttribute(b, &user, lduser.IPAttribute, user.GetIP(), &privateAttrs)
	uf.maybeStringAttribute(b, &user, lduser.CountryAttribute, user.GetCountry(), &privateAttrs)
	uf.maybeStringAttribute(b, &user, lduser.EmailAttribute, user.GetEmail(), &privateAttrs)
	uf.maybeStringAttribute(b, &user, lduser.LastNameAttribute, user.GetLastName(), &privateAttrs)
	uf.maybeStringAttribute(b, &user, lduser.FirstNameAttribute, user.GetFirstName(), &privateAttrs)
	uf.maybeStringAttribute(b, &user, lduser.AvatarAttribute, user.GetAvatar(), &privateAttrs)
	uf.maybeStringAttribute(b, &user, lduser.NameAttribute, user.GetName(), &privateAttrs)

	if anon, hasAnon := user.GetAnonymousOptional(); hasAnon {
		b.WriteName(string(lduser.AnonymousAttribute))
		b.WriteBool(anon)
	}

	custom := user.GetAllCustom()
	wroteCustom := false
	if custom.Count() > 0 {
		custom.Enumerate(func(i int, key string, value ldvalue.Value) bool {
			if uf.isPrivate(&user, lduser.UserAttribute(key)) {
				privateAttrs = append(privateAttrs, key)
			} else {
				if !wroteCustom {
					b.WriteName("custom")
					b.BeginObject()
					wroteCustom = true
				}
				b.WriteName(key)
				value.WriteToJSONBuffer(b)
			}
			return true
		})
	}
	if wroteCustom {
		b.EndObject()
	}

	if user.AlreadyFilteredAttributes != nil {
		privateAttrs = user.AlreadyFilteredAttributes
	}
	if len(privateAttrs) > 0 {
		b.WriteName("privateAttrs")
		b.BeginArray()
		for _, a := range privateAttrs {
			b.WriteString(a)
		}
		b.EndArray()
	}

	b.EndObject()
}

func (uf *userFilter) maybeStringAttribute(
	b *jsonstream.JSONBuffer,
	user *EventUser,
	attrName lduser.UserAttribute,
	value ldvalue.OptionalString,
	privateAttrs *[]string) {
	if s, ok := value.Get(); ok {
		if uf.isPrivate(user, attrName) {
			*privateAttrs = append(*privateAttrs, string(attrName))
		} else {
			b.WriteName(string(attrName))
			b.WriteString(s)
		}
	}
}

func (uf *userFilter) isPrivate(user *EventUser, attrName lduser.UserAttribute) bool {
	// If alreadyFiltered is true, it means this is user data that has already gone through the
	// attribute filtering logic, so the private attribute values have already been removed and their
	// names are in user.AlreadyFilteredAttributes. This happens when Relay receives event data from
	// the PHP SDK. In this case, we do not need to repeat the filtering logic and we do not support
	// re-filtering with a different private attribute configuration.
	if user.AlreadyFilteredAttributes != nil {
		return false
	}
	if uf.allAttributesPrivate || user.IsPrivateAttribute(attrName) {
		return true
	}
	for _, a := range uf.globalPrivateAttributes {
		if a == attrName {
			return true
		}
	}
	return false
}
