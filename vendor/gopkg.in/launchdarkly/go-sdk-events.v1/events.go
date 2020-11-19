package ldevents

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldreason"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldtime"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

// EventUser is a combination of the standard User struct with additional information that may
// be relevant outside of the standard SDK event generation context.
//
// Specifically, the AlreadyFilteredAttributes property is used by ld-relay when it has received
// event data from the PHP SDK that has not gone through the summarization process, but *has*
// gone through private attribute filtering. Since that filtering normally only happens as part
// of event production and is not actually a property of the user, the lduser builder does not
// have a way to embed that information in the user so it is added here separately.
type EventUser struct {
	lduser.User
	// AlreadyFilteredAttributes is a list of private attribute names that were already removed
	// from the user before generating the event. If this is non-nil, the usual attribute
	// filtering logic will be skipped and this list will be passed on unchanged in the
	// privateAttrs property of the output user.
	AlreadyFilteredAttributes []string
}

// User is a convenience function to convert lduser.User to EventUser.
func User(baseUser lduser.User) EventUser {
	return EventUser{baseUser, nil}
}

// Event represents an analytics event generated by the client, which will be passed to
// the EventProcessor.  The event data that the EventProcessor actually sends to LaunchDarkly
// may be slightly different.
type Event interface {
	GetBase() BaseEvent
}

// BaseEvent provides properties common to all events.
type BaseEvent struct {
	CreationDate ldtime.UnixMillisecondTime
	User         EventUser
}

// FeatureRequestEvent is generated by evaluating a feature flag or one of a flag's prerequisites.
type FeatureRequestEvent struct {
	BaseEvent
	Key                  string
	Variation            ldvalue.OptionalInt
	Value                ldvalue.Value
	Default              ldvalue.Value
	Version              ldvalue.OptionalInt
	PrereqOf             ldvalue.OptionalString
	Reason               ldreason.EvaluationReason
	TrackEvents          bool
	Debug                bool
	DebugEventsUntilDate ldtime.UnixMillisecondTime
}

// CustomEvent is generated by calling the client's Track method.
type CustomEvent struct {
	BaseEvent
	Key         string
	Data        ldvalue.Value
	HasMetric   bool
	MetricValue float64
}

// IdentifyEvent is generated by calling the client's Identify method.
type IdentifyEvent struct {
	BaseEvent
}

// indexEvent is generated internally to capture user details from other events. It is an implementation
// detail of DefaultEventProcessor, so it is not exported.
type indexEvent struct {
	BaseEvent
}

// EventFactory is a configurable factory for event objects.
type EventFactory struct {
	includeReasons bool
	timeFn         func() ldtime.UnixMillisecondTime
}

// NewEventFactory creates an EventFactory.
//
// The includeReasons parameter is true if evaluation events should always include the EvaluationReason (this is
// used by the SDK when one of the "VariationDetail" methods is called). The timeFn parameter is normally nil but
// can be used to instrument the EventFactory with a source of time data other than the standard clock.
//
// The isExperimentFn parameter is necessary to provide the additional experimentation behavior that is
func NewEventFactory(includeReasons bool, timeFn func() ldtime.UnixMillisecondTime) EventFactory {
	if timeFn == nil {
		timeFn = ldtime.UnixMillisNow
	}
	return EventFactory{includeReasons, timeFn}
}

// NewUnknownFlagEvent creates an evaluation event for a missing flag.
func (f EventFactory) NewUnknownFlagEvent(
	key string,
	user EventUser,
	defaultVal ldvalue.Value,
	reason ldreason.EvaluationReason,
) FeatureRequestEvent {
	fre := FeatureRequestEvent{
		BaseEvent: BaseEvent{
			CreationDate: f.timeFn(),
			User:         user,
		},
		Key:     key,
		Value:   defaultVal,
		Default: defaultVal,
	}
	if f.includeReasons {
		fre.Reason = reason
	}
	return fre
}

// NewEvalEvent creates an evaluation event for an existing flag.
func (f EventFactory) NewEvalEvent(
	flagProps FlagEventProperties,
	user EventUser,
	detail ldreason.EvaluationDetail,
	defaultVal ldvalue.Value,
	prereqOf string,
) FeatureRequestEvent {
	requireExperimentData := flagProps.IsExperimentationEnabled(detail.Reason)
	fre := FeatureRequestEvent{
		BaseEvent: BaseEvent{
			CreationDate: f.timeFn(),
			User:         user,
		},
		Key:                  flagProps.GetKey(),
		Version:              ldvalue.NewOptionalInt(flagProps.GetVersion()),
		Variation:            detail.VariationIndex,
		Value:                detail.Value,
		Default:              defaultVal,
		TrackEvents:          requireExperimentData || flagProps.IsFullEventTrackingEnabled(),
		DebugEventsUntilDate: flagProps.GetDebugEventsUntilDate(),
	}
	if f.includeReasons || requireExperimentData {
		fre.Reason = detail.Reason
	}
	if prereqOf != "" {
		fre.PrereqOf = ldvalue.NewOptionalString(prereqOf)
	}
	return fre
}

// GetBase returns the BaseEvent
func (evt FeatureRequestEvent) GetBase() BaseEvent {
	return evt.BaseEvent
}

// NewCustomEvent creates a new custom event.
func (f EventFactory) NewCustomEvent(
	key string,
	user EventUser,
	data ldvalue.Value,
	withMetric bool,
	metricValue float64,
) CustomEvent {
	ce := CustomEvent{
		BaseEvent: BaseEvent{
			CreationDate: f.timeFn(),
			User:         user,
		},
		Key:         key,
		Data:        data,
		HasMetric:   withMetric,
		MetricValue: metricValue,
	}
	return ce
}

// GetBase returns the BaseEvent
func (evt CustomEvent) GetBase() BaseEvent {
	return evt.BaseEvent
}

// NewIdentifyEvent constructs a new identify event, but does not send it. Typically, Identify should be
// used to both create the event and send it to LaunchDarkly.
func (f EventFactory) NewIdentifyEvent(user EventUser) IdentifyEvent {
	return IdentifyEvent{
		BaseEvent: BaseEvent{
			CreationDate: f.timeFn(),
			User:         user,
		},
	}
}

// GetBase returns the BaseEvent
func (evt IdentifyEvent) GetBase() BaseEvent {
	return evt.BaseEvent
}

// GetBase returns the BaseEvent
func (evt indexEvent) GetBase() BaseEvent {
	return evt.BaseEvent
}