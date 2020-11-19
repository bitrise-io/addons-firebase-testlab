package ldmodel

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldreason"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldtime"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

// FeatureFlag describes an individual feature flag.
//
// The fields of this struct are exported for use by LaunchDarkly internal components. Application code
// should normally not reference FeatureFlag fields directly; flag data normally comes from LaunchDarkly
// SDK endpoints in JSON form and can be deserialized using the DataModelSerialization interface.
type FeatureFlag struct {
	// Key is the unique string key of the feature flag.
	Key string
	// On is true if targeting is turned on for this flag.
	//
	// If On is false, the evaluator always uses OffVariation and ignores all other fields.
	On bool
	// Prerequisites is a list of feature flag conditions that are prerequisites for this flag.
	//
	// If any prerequisite is not met, the flag behaves as if targeting is turned off.
	Prerequisites []Prerequisite
	// Targets contains sets of individually targeted users.
	//
	// Targets take precedence over Rules: if a user is matched by any Target, the Rules are ignored.
	// Targets are ignored if targeting is turned off.
	Targets []Target
	// Rules is a list of rules that may match a user.
	//
	// If a user is matched by a Rule, all subsequent Rules in the list are skipped. Rules are ignored
	// if targeting is turned off.
	Rules []FlagRule
	// Fallthrough defines the flag's behavior if targeting is turned on but the user is not matched
	// by any Target or Rule.
	Fallthrough VariationOrRollout
	// OffVariation specifies the variation index to use if targeting is turned off.
	//
	// If this is undefined (ldvalue.OptionalInt{}), Evaluate returns undefined for the variation
	// index and ldvalue.Null() for the value.
	OffVariation ldvalue.OptionalInt
	// Variations is the list of all allowable variations for this flag. The variation index in a
	// Target or Rule is a zero-based index to this list.
	Variations []ldvalue.Value
	// ClientSideAvailability indicates whether a flag is available using each of the client-side
	// authentication methods.
	ClientSideAvailability ClientSideAvailability
	// Salt is a randomized value assigned to this flag when it is created.
	//
	// The hash function used for calculating percentage rollouts uses this as a salt to ensure that
	// rollouts are consistent within each flag but not predictable from one flag to another.
	Salt string
	// TrackEvents is used internally by the SDK analytics event system.
	//
	// This field is true if the current LaunchDarkly account has data export enabled, and has turned on
	// the "send detailed event information for this flag" option for this flag. This tells the SDK to
	// send full event data for each flag evaluation, rather than only aggregate data in a summary event.
	//
	// The go-server-sdk-evaluation package does not implement that behavior; it is only in the data
	// model for use by the SDK.
	TrackEvents bool
	// TrackEventsFallthrough is used internally by the SDK analytics event system.
	//
	// This field is true if the current LaunchDarkly account has experimentation enabled, has associated
	// this flag with an experiment, and has enabled "default rule" for the experiment. This tells the
	// SDK to send full event data for any evaluation where this flag had targeting turned on but the
	// user did not match any targets or rules.
	//
	// The go-server-sdk-evaluation package does not implement that behavior; it is only in the data
	// model for use by the SDK.
	TrackEventsFallthrough bool
	// DebugEventsUntilDate is used internally by the SDK analytics event system.
	//
	// This field is non-zero if debugging for this flag has been turned on temporarily in the
	// LaunchDarkly dashboard. Debugging always is for a limited time, so the field specifies a Unix
	// millisecond timestamp when this mode should expire. Until then, the SDK will send full event data
	// for each evaluation of this flag.
	//
	// The go-server-sdk-evaluation package does not implement that behavior; it is only in the data
	// model for use by the SDK.
	DebugEventsUntilDate ldtime.UnixMillisecondTime
	// Version is an integer that is incremented by LaunchDarkly every time the configuration of the flag is
	// changed.
	Version int
	// Deleted is true if this is not actually a feature flag but rather a placeholder (tombstone) for a
	// deleted flag. This is only relevant in data store implementations. The SDK does not evaluate
	// deleted flags.
	Deleted bool
}

// GetKey returns the string key for the flag.
//
// This method exists in order to conform to interfaces used internally by the SDK.
func (f *FeatureFlag) GetKey() string {
	return f.Key
}

// GetVersion returns the version of the flag.
//
// This method exists in order to conform to interfaces used internally by the SDK.
func (f *FeatureFlag) GetVersion() int {
	return f.Version
}

// IsFullEventTrackingEnabled returns true if the flag has been configured to always generate detailed event data.
//
// This method exists in order to conform to interfaces used internally by the SDK
// (go-sdk-events.v1/FlagEventProperties). It simply returns TrackEvents.
func (f *FeatureFlag) IsFullEventTrackingEnabled() bool {
	return f.TrackEvents
}

// GetDebugEventsUntilDate returns zero normally, but if event debugging has been temporarily enabled for the flag,
// it returns the time at which debugging mode should expire.
//
// This method exists in order to conform to interfaces used internally by the SDK
// (go-sdk-events.v1/FlagEventProperties). It simply returns DebugEventsUntilDate.
func (f *FeatureFlag) GetDebugEventsUntilDate() ldtime.UnixMillisecondTime {
	return f.DebugEventsUntilDate
}

// IsExperimentationEnabled returns true if, based on the EvaluationReason returned by the flag evaluation, an event for
// that evaluation should have full tracking enabled and always report the reason even if the application didn't
// explicitly request this. For instance, this is true if a rule was matched that had tracking enabled for that specific
// rule.
//
// This differs from IsFullEventTrackingEnabled() in that it is dependent on the result of a specific evaluation; also,
// IsFullEventTrackingEnabled() being true does not imply that the event should always contain a reason, whereas
// IsExperimentationEnabled() being true does force the reason to be included.
//
// This method exists in order to conform to interfaces used internally by the SDK
// (go-sdk-events.v1/FlagEventProperties).
func (f *FeatureFlag) IsExperimentationEnabled(reason ldreason.EvaluationReason) bool {
	switch reason.GetKind() {
	case ldreason.EvalReasonFallthrough:
		return f.TrackEventsFallthrough
	case ldreason.EvalReasonRuleMatch:
		i := reason.GetRuleIndex()
		if i >= 0 && i < len(f.Rules) {
			return f.Rules[i].TrackEvents
		}
	}
	return false
}

// FlagRule describes a single rule within a feature flag.
//
// A rule consists of a set of ANDed matching conditions (Clause) for a user, along with either a fixed
// variation or a set of rollout percentages to use if the user matches all of the clauses.
type FlagRule struct {
	// VariationRollout properties for a FlagRule define what variation to return if the user matches
	// this rule.
	VariationOrRollout
	// ID is a randomized identifier assigned to each rule when it is created.
	//
	// This is used to populate the RuleID property of ldreason.EvaluationReason.
	ID string
	// Clauses is a list of test conditions that make up the rule. These are ANDed: every Clause must
	// match in order for the FlagRule to match.
	Clauses []Clause
	// TrackEvents is used internally by the SDK analytics event system.
	//
	// This field is true if the current LaunchDarkly account has experimentation enabled, has associated
	// this flag with an experiment, and has enabled this rule for the experiment. This tells the SDK to
	// send full event data for any evaluation that matches this rule.
	//
	// The go-server-sdk-evaluation package does not implement that behavior; it is only in the data
	// model for use by the SDK.
	TrackEvents bool
}

// VariationOrRollout desscribes either a fixed variation or a percentage rollout.
//
// There is a VariationOrRollout for every FlagRule, and also one in FeatureFlag.Fallthrough which is
// used if no rules match.
//
// Invariant: one of the variation or rollout must be non-nil.
type VariationOrRollout struct {
	// Variation specifies the index of the variation to return. It is undefined (ldvalue.OptionalInt{})
	// if no specific variation is defined.
	Variation ldvalue.OptionalInt
	// Rollout specifies a percentage rollout to be used instead of a specific variation. A rollout is
	// only defined if it has a non-empty Variations list.
	Rollout Rollout
}

// Rollout describes how users will be bucketed into variations during a percentage rollout.
type Rollout struct {
	// Variations is a list of the variations in the percentage rollout and what percentage of users
	// to include in each.
	//
	// The Weight values of all elements in this list should add up to 100000 (100%). If they do not,
	// the last element in the list will behave as if it includes any leftover percentage (that is, if
	// the weights are [1000, 1000, 1000] they will be treated as if they were [1000, 1000, 99000]).
	Variations []WeightedVariation
	// BucketBy specifies which user attribute should be used to distinguish between users in a rollout.
	//
	// The default (when BucketBy is empty) is lduser.KeyAttribute, the user's primary key. If you wish to
	// treat users with different keys as the same for rollout purposes as long as they have the same
	// "country" attribute, you would set this to "country" (lduser.CountryAttribute).
	//
	// Rollouts always take the user's "secondary key" attribute into account as well if the user has one.
	BucketBy lduser.UserAttribute
}

// Clause describes an individual cluuse within a FlagRule or SegmentRule.
type Clause struct {
	// Attribute specifies the user attribute that is being tested.
	//
	// This is required for all Operator types except SegmentMatch.
	//
	// If the user's value for this attribute is a JSON array, then the test specified in the Clause is
	// repeated for each value in the array until a match is found or there are no more values.
	Attribute lduser.UserAttribute
	// Op specifies the type of test to perform.
	Op Operator
	// Values is a list of values to be compared to the user attribute.
	//
	// This is interpreted as an OR: if the user attribute matches any of these values with the specified
	// operator, the Clause matches the user.
	//
	// In the special case where Op is OperatorSegmentMtach, there should only be a single Value, which
	// must be a string: the key of the user segment.
	//
	// If the user does not have a value for the specified attribute, the Values are ignored and the
	// Clause is always treated as a non-match.
	Values []ldvalue.Value
	// Negate is true if the specified Operator should be inverted.
	//
	// For instance, this would cause OperatorIn to mean "not equal" rather than "equal". Note that if no
	// tests are performed for this Clause because the user does not have a value for the specified
	// attribute, then Negate will not come into effect (the Clause will just be treated as a non-match).
	Negate bool
	// preprocessed is created by PreprocessFlag() to speed up clause evaluation in scenarios like
	// regex matching.
	preprocessed clausePreprocessedData
}

// WeightedVariation describes a fraction of users who will receive a specific variation.
type WeightedVariation struct {
	// Variation is the index of the variation to be returned if the user is in this bucket. This is
	// always a real variation index; it cannot be undefined.
	Variation int
	// Weight is the proportion of users who should go into this bucket, as an integer from 0 to 100000.
	Weight int
}

// Target describes a set of users who will receive a specific variation.
type Target struct {
	// Values is the set of user keys included in this Target.
	Values []string
	// Variation is the index of the variation to be returned if the user matches one of these keys. This
	// is always a real variation index; it cannot be undefined.
	Variation int
	// preprocessedData is created by PreprocessFlag() to speed up target matching.
	preprocessed targetPreprocessedData
}

// Prerequisite describes a requirement that another feature flag return a specific variation.
//
// A prerequisite condition is met if the specified prerequisite flag has targeting turned on and
// returns the specified variation.
type Prerequisite struct {
	// Key is the unique key of the feature flag to be evaluated as a prerequisite.
	Key string
	// Variation is the index of the variation that the prerequisite flag must return in order for
	// the prerequisite condition to be met. If the prerequisite flag has targeting turned on, then
	// the condition is not met even if the flag's OffVariation matches this value. This is always a
	// real variation index; it cannot be undefined.
	Variation int
}

// ClientSideAvailability describes whether a flag is available to client-side SDKs.
//
// This field can be used by a server-side client to determine whether to include an individual flag in
// bootstrapped set of flag data (see https://docs.launchdarkly.com/sdk/client-side/javascript#bootstrapping).
type ClientSideAvailability struct {
	// UsingMobileKey indicates that this flag is available to clients using the mobile key for authorization
	// (includes most desktop and mobile clients).
	UsingMobileKey bool
	// UsingEnvironmentID indicates that this flag is available to clients using the environment id to identify an
	// environment (includes client-side javascript clients).
	UsingEnvironmentID bool
	// Explicit is true if, when serializing this flag, all of the ClientSideAvailability properties should
	// be included. If it is false, then an older schema is used in which this object is entirely omitted,
	// UsingEnvironmentID is stored in a deprecated property, and UsingMobileKey is assumed to be true.
	//
	// This field exists to ensure that flag representations remain consistent when sent and received
	// even though the clientSideAvailability property may not be present in the JSON data. It is false
	// if the flag was deserialized from an older JSON schema that did not include that property.
	//
	// Similarly, when deserializing a flag, if it used the older schema then Explicit will be false and
	// UsingMobileKey will be true.
	Explicit bool
}
