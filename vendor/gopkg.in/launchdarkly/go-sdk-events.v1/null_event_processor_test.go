package ldevents

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldreason"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

func TestNullEventProcessor(t *testing.T) {
	// Just verifies that these methods don't panic
	n := NewNullEventProcessor()
	n.RecordFeatureRequestEvent(defaultEventFactory.NewUnknownFlagEvent("x", epDefaultUser, ldvalue.Null(),
		ldreason.EvaluationReason{}))
	n.RecordIdentifyEvent(defaultEventFactory.NewIdentifyEvent(epDefaultUser))
	n.RecordCustomEvent(defaultEventFactory.NewCustomEvent("x", epDefaultUser, ldvalue.Null(), false, 0))
	n.Flush()
	require.NoError(t, n.Close())
}
