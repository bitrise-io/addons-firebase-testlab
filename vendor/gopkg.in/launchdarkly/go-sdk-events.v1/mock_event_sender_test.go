package ldevents

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

type mockEventSender struct {
	events             []ldvalue.Value
	diagnosticEvents   []ldvalue.Value
	eventsCh           chan ldvalue.Value
	diagnosticEventsCh chan ldvalue.Value
	payloadCount       int
	result             EventSenderResult
	gateCh             <-chan struct{}
	waitingCh          chan<- struct{}
	lock               sync.Mutex
}

func newMockEventSender() *mockEventSender {
	return &mockEventSender{
		eventsCh:           make(chan ldvalue.Value, 100),
		diagnosticEventsCh: make(chan ldvalue.Value, 100),
		result:             EventSenderResult{Success: true},
	}
}

func (ms *mockEventSender) SendEventData(kind EventDataKind, data []byte, eventCount int) EventSenderResult {
	var jsonData ldvalue.Value
	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		panic(err)
	}

	ms.lock.Lock()
	if kind == DiagnosticEventDataKind {
		ms.diagnosticEvents = append(ms.diagnosticEvents, jsonData)
		ms.diagnosticEventsCh <- jsonData
	} else {
		jsonData.Enumerate(func(i int, k string, v ldvalue.Value) bool {
			ms.events = append(ms.events, v)
			ms.eventsCh <- v
			return true
		})
		ms.payloadCount++
	}
	gateCh, waitingCh := ms.gateCh, ms.waitingCh
	result := ms.result
	ms.lock.Unlock()

	if gateCh != nil {
		// instrumentation used for TestEventsAreKeptInBufferIfAllFlushWorkersAreBusy
		waitingCh <- struct{}{}
		<-gateCh
	}

	return result
}

func (ms *mockEventSender) setGate(gateCh <-chan struct{}, waitingCh chan<- struct{}) {
	ms.lock.Lock()
	ms.gateCh = gateCh
	ms.waitingCh = waitingCh
	ms.lock.Unlock()
}

func (ms *mockEventSender) getPayloadCount() int {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	return ms.payloadCount
}

func (ms *mockEventSender) awaitEvent(t *testing.T) ldvalue.Value {
	return ms.awaitEventCh(t, ms.eventsCh)
}

func (ms *mockEventSender) awaitDiagnosticEvent(t *testing.T) ldvalue.Value {
	return ms.awaitEventCh(t, ms.diagnosticEventsCh)
}

func (ms *mockEventSender) awaitEventCh(t *testing.T, ch <-chan ldvalue.Value) ldvalue.Value {
	select {
	case e := <-ch:
		return e
	case <-time.After(time.Second):
		require.Fail(t, "expected an event but did not receive one")
	}
	return ldvalue.Null()
}

func (ms *mockEventSender) assertNoMoreEvents(t *testing.T) {
	require.Len(t, ms.eventsCh, 0)
}

func (ms *mockEventSender) assertNoMoreDiagnosticEvents(t *testing.T) {
	require.Len(t, ms.diagnosticEventsCh, 0)
}
